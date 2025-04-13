package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	goopenai "github.com/sashabaranov/go-openai"

	"github.com/danielmiessler/fabric/common"
	"github.com/danielmiessler/fabric/plugins/ai"
	"github.com/danielmiessler/fabric/plugins/db/fsdb"
	"github.com/danielmiessler/fabric/plugins/strategy"
	"github.com/danielmiessler/fabric/plugins/template"
)

const NoSessionPatternUserMessages = "no session, pattern or user messages provided"

type Chatter struct {
	db *fsdb.Db

	Stream bool
	DryRun bool

	model              string
	modelContextLength int
	vendor             ai.Vendor
	strategy           string
}

// Send processes a chat request and applies any file changes if using the create_coding_feature pattern
func (o *Chatter) Send(request *common.ChatRequest, opts *common.ChatOptions) (session *fsdb.Session, err error) {
	if session, err = o.BuildSession(request, opts.Raw); err != nil {
		return
	}

	vendorMessages := session.GetVendorMessages()
	if len(vendorMessages) == 0 {
		if session.Name != "" {
			err = o.db.Sessions.SaveSession(session)
			if err != nil {
				return
			}
		}
		err = fmt.Errorf("no messages provided")
		return
	}

	if opts.Model == "" {
		opts.Model = o.model
	}

	if opts.ModelContextLength == 0 {
		opts.ModelContextLength = o.modelContextLength
	}

	message := ""

	if o.Stream {
		channel := make(chan string)
		go func() {
			if streamErr := o.vendor.SendStream(session.GetVendorMessages(), opts, channel); streamErr != nil {
				channel <- streamErr.Error()
			}
		}()

		for response := range channel {
			message += response
			fmt.Print(response)
		}
	} else {
		if message, err = o.vendor.Send(context.Background(), session.GetVendorMessages(), opts); err != nil {
			return
		}
	}

	if message == "" {
		session = nil
		err = fmt.Errorf("empty response")
		return
	}

	// Process file changes if using the create_coding_feature pattern
	if request.PatternName == "create_coding_feature" {
		// Look for file changes in the response
		summary, fileChanges, parseErr := common.ParseFileChanges(message)
		if parseErr != nil {
			fmt.Printf("Warning: Failed to parse file changes: %v\n", parseErr)
		} else if len(fileChanges) > 0 {
			// Get the project root - use the current directory
			projectRoot, err := os.Getwd()
			if err != nil {
				fmt.Printf("Warning: Failed to get current directory: %v\n", err)
				// Continue without applying changes
			} else {
				if applyErr := common.ApplyFileChanges(projectRoot, fileChanges); applyErr != nil {
					fmt.Printf("Warning: Failed to apply file changes: %v\n", applyErr)
				} else {
					fmt.Println("Successfully applied file changes.")
					fmt.Printf("You can review the changes with 'git diff' if you're using git.\n\n")
				}
			}
		}
		message = summary
	}

	session.Append(&goopenai.ChatCompletionMessage{Role: goopenai.ChatMessageRoleAssistant, Content: message})

	if session.Name != "" {
		err = o.db.Sessions.SaveSession(session)
	}
	return
}

func (o *Chatter) BuildSession(request *common.ChatRequest, raw bool) (session *fsdb.Session, err error) {
	// If a session name is provided, retrieve it from the database
	if request.SessionName != "" {
		var sess *fsdb.Session
		if sess, err = o.db.Sessions.Get(request.SessionName); err != nil {
			err = fmt.Errorf("could not find session %s: %v", request.SessionName, err)
			return
		}
		session = sess
	} else {
		session = &fsdb.Session{}
	}

	if request.Meta != "" {
		session.Append(&goopenai.ChatCompletionMessage{Role: common.ChatMessageRoleMeta, Content: request.Meta})
	}

	// if a context name is provided, retrieve it from the database
	var contextContent string
	if request.ContextName != "" {
		var ctx *fsdb.Context
		if ctx, err = o.db.Contexts.Get(request.ContextName); err != nil {
			err = fmt.Errorf("could not find context %s: %v", request.ContextName, err)
			return
		}
		contextContent = ctx.Content
	}

	// Process any template variables in the message content (user input)
	// Double curly braces {{variable}} indicate template substitution
	// Ensure we have a message before processing, other wise we'll get an error when we pass to pattern.go
	if request.Message == nil {
		request.Message = &goopenai.ChatCompletionMessage{
			Role:    goopenai.ChatMessageRoleUser,
			Content: " ",
		}
	}

	// Now we know request.Message is not nil, process template variables
	if request.InputHasVars {
		request.Message.Content, err = template.ApplyTemplate(request.Message.Content, request.PatternVariables, "")
		if err != nil {
			return nil, err
		}
	}

	var patternContent string
	if request.PatternName != "" {
		pattern, err := o.db.Patterns.GetApplyVariables(request.PatternName, request.PatternVariables, request.Message.Content)
		// pattern will now contain user input, and all variables will be resolved, or errored

		if err != nil {
			return nil, fmt.Errorf("could not get pattern %s: %v", request.PatternName, err)
		}
		patternContent = pattern.Pattern
	}

	systemMessage := strings.TrimSpace(contextContent) + strings.TrimSpace(patternContent)

	// Apply strategy if specified
	if request.StrategyName != "" {
		strategy, err := strategy.LoadStrategy(request.StrategyName)
		if err != nil {
			return nil, fmt.Errorf("could not load strategy %s: %v", request.StrategyName, err)
		}
		if strategy != nil && strategy.Prompt != "" {
			// prepend the strategy prompt to the system message
			systemMessage = fmt.Sprintf("%s\n%s", strategy.Prompt, systemMessage)
		}
	}

	if request.Language != "" {
		systemMessage = fmt.Sprintf("%s. Please use the language '%s' for the output.", systemMessage, request.Language)
	}

	if raw {
		if request.Message != nil {
			if systemMessage != "" {
				request.Message.Content = systemMessage
				// system contains pattern which contains user input
			}
		} else {
			if systemMessage != "" {
				request.Message = &goopenai.ChatCompletionMessage{Role: goopenai.ChatMessageRoleSystem, Content: systemMessage}
			}
		}
	} else {
		if systemMessage != "" {
			session.Append(&goopenai.ChatCompletionMessage{Role: goopenai.ChatMessageRoleSystem, Content: systemMessage})
		}
	}

	if request.Message != nil {
		session.Append(request.Message)
	}

	if session.IsEmpty() {
		session = nil
		err = errors.New(NoSessionPatternUserMessages)
	}
	return
}
