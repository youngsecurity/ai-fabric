package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"
	goopenai "github.com/sashabaranov/go-openai"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"

	"github.com/danielmiessler/fabric/common"
)

// Flags create flags struct. the users flags go into this, this will be passed to the chat struct in cli
type Flags struct {
	Pattern                         string            `short:"p" long:"pattern" yaml:"pattern" description:"Choose a pattern from the available patterns" default:""`
	PatternVariables                map[string]string `short:"v" long:"variable" description:"Values for pattern variables, e.g. -v=#role:expert -v=#points:30"`
	Context                         string            `short:"C" long:"context" description:"Choose a context from the available contexts" default:""`
	Session                         string            `long:"session" description:"Choose a session from the available sessions"`
	Attachments                     []string          `short:"a" long:"attachment" description:"Attachment path or URL (e.g. for OpenAI image recognition messages)"`
	Setup                           bool              `short:"S" long:"setup" description:"Run setup for all reconfigurable parts of fabric"`
	Temperature                     float64           `short:"t" long:"temperature" yaml:"temperature" description:"Set temperature" default:"0.7"`
	TopP                            float64           `short:"T" long:"topp" yaml:"topp" description:"Set top P" default:"0.9"`
	Stream                          bool              `short:"s" long:"stream" yaml:"stream" description:"Stream"`
	PresencePenalty                 float64           `short:"P" long:"presencepenalty" yaml:"presencepenalty" description:"Set presence penalty" default:"0.0"`
	Raw                             bool              `short:"r" long:"raw" yaml:"raw" description:"Use the defaults of the model without sending chat options (like temperature etc.) and use the user role instead of the system role for patterns."`
	FrequencyPenalty                float64           `short:"F" long:"frequencypenalty" yaml:"frequencypenalty" description:"Set frequency penalty" default:"0.0"`
	ListPatterns                    bool              `short:"l" long:"listpatterns" description:"List all patterns"`
	ListAllModels                   bool              `short:"L" long:"listmodels" description:"List all available models"`
	ListAllContexts                 bool              `short:"x" long:"listcontexts" description:"List all contexts"`
	ListAllSessions                 bool              `short:"X" long:"listsessions" description:"List all sessions"`
	UpdatePatterns                  bool              `short:"U" long:"updatepatterns" description:"Update patterns"`
	Message                         string            `hidden:"true" description:"Messages to send to chat"`
	Copy                            bool              `short:"c" long:"copy" description:"Copy to clipboard"`
	Model                           string            `short:"m" long:"model" yaml:"model" description:"Choose model"`
	ModelContextLength              int               `long:"modelContextLength" yaml:"modelContextLength" description:"Model context length (only affects ollama)"`
	Output                          string            `short:"o" long:"output" description:"Output to file" default:""`
	OutputSession                   bool              `long:"output-session" description:"Output the entire session (also a temporary one) to the output file"`
	LatestPatterns                  string            `short:"n" long:"latest" description:"Number of latest patterns to list" default:"0"`
	ChangeDefaultModel              bool              `short:"d" long:"changeDefaultModel" description:"Change default model"`
	YouTube                         string            `short:"y" long:"youtube" description:"YouTube video or play list \"URL\" to grab transcript, comments from it and send to chat or print it put to the console and store it in the output file"`
	YouTubePlaylist                 bool              `long:"playlist" description:"Prefer playlist over video if both ids are present in the URL"`
	YouTubeTranscript               bool              `long:"transcript" description:"Grab transcript from YouTube video and send to chat (it is used per default)."`
	YouTubeTranscriptWithTimestamps bool              `long:"transcript-with-timestamps" description:"Grab transcript from YouTube video with timestamps and send to chat"`
	YouTubeComments                 bool              `long:"comments" description:"Grab comments from YouTube video and send to chat"`
	YouTubeMetadata                 bool              `long:"metadata" description:"Output video metadata"`
	Language                        string            `short:"g" long:"language" description:"Specify the Language Code for the chat, e.g. -g=en -g=zh" default:""`
	ScrapeURL                       string            `short:"u" long:"scrape_url" description:"Scrape website URL to markdown using Jina AI"`
	ScrapeQuestion                  string            `short:"q" long:"scrape_question" description:"Search question using Jina AI"`
	Seed                            int               `short:"e" long:"seed" yaml:"seed" description:"Seed to be used for LMM generation"`
	WipeContext                     string            `short:"w" long:"wipecontext" description:"Wipe context"`
	WipeSession                     string            `short:"W" long:"wipesession" description:"Wipe session"`
	PrintContext                    string            `long:"printcontext" description:"Print context"`
	PrintSession                    string            `long:"printsession" description:"Print session"`
	HtmlReadability                 bool              `long:"readability" description:"Convert HTML input into a clean, readable view"`
	InputHasVars                    bool              `long:"input-has-vars" description:"Apply variables to user input"`
	DryRun                          bool              `long:"dry-run" description:"Show what would be sent to the model without actually sending it"`
	Serve                           bool              `long:"serve" description:"Serve the Fabric Rest API"`
	ServeOllama                     bool              `long:"serveOllama" description:"Serve the Fabric Rest API with ollama endpoints"`
	ServeAddress                    string            `long:"address" description:"The address to bind the REST API" default:":8080"`
	Config                          string            `long:"config" description:"Path to YAML config file"`
	Version                         bool              `long:"version" description:"Print current version"`
	ListExtensions                  bool              `long:"listextensions" description:"List all registered extensions"`
	AddExtension                    string            `long:"addextension" description:"Register a new extension from config file path"`
	RemoveExtension                 string            `long:"rmextension" description:"Remove a registered extension by name"`
}

var debug = false

func Debugf(format string, a ...interface{}) {
	if debug {
		fmt.Printf("DEBUG: "+format, a...)
	}
}

// Init Initialize flags. returns a Flags struct and an error
func Init() (ret *Flags, err error) {
	// Track which yaml-configured flags were set on CLI
	usedFlags := make(map[string]bool)
	yamlArgsScan := os.Args[1:]

	// Get list of fields that have yaml tags, could be in yaml config
	yamlFields := make(map[string]bool)
	t := reflect.TypeOf(Flags{})
	for i := 0; i < t.NumField(); i++ {
		if yamlTag := t.Field(i).Tag.Get("yaml"); yamlTag != "" {
			yamlFields[yamlTag] = true
			//Debugf("Found yaml-configured field: %s\n", yamlTag)
		}
	}

	// Scan args for that are provided by cli and might be in yaml
	for _, arg := range yamlArgsScan {
		if strings.HasPrefix(arg, "--") {
			flag := strings.TrimPrefix(arg, "--")
			if i := strings.Index(flag, "="); i > 0 {
				flag = flag[:i]
			}
			if yamlFields[flag] {
				usedFlags[flag] = true
				Debugf("CLI flag used: %s\n", flag)
			}
		}
	}

	// Parse CLI flags first
	ret = &Flags{}
	parser := flags.NewParser(ret, flags.Default)
	var args []string
	if args, err = parser.Parse(); err != nil {
		return nil, err
	}

	// If config specified, load and apply YAML for unused flags
	if ret.Config != "" {
		yamlFlags, err := loadYAMLConfig(ret.Config)
		if err != nil {
			return nil, err
		}

		// Apply YAML values where CLI flags weren't used
		flagsVal := reflect.ValueOf(ret).Elem()
		yamlVal := reflect.ValueOf(yamlFlags).Elem()
		flagsType := flagsVal.Type()

		for i := 0; i < flagsType.NumField(); i++ {
			field := flagsType.Field(i)
			if yamlTag := field.Tag.Get("yaml"); yamlTag != "" {
				if !usedFlags[yamlTag] {
					flagField := flagsVal.Field(i)
					yamlField := yamlVal.Field(i)
					if flagField.CanSet() {
						if yamlField.Type() != flagField.Type() {
							if err := assignWithConversion(flagField, yamlField); err != nil {
								Debugf("Type conversion failed for %s: %v\n", yamlTag, err)
								continue
							}
						} else {
							flagField.Set(yamlField)
						}
						Debugf("Applied YAML value for %s: %v\n", yamlTag, yamlField.Interface())
					}
				}
			}
		}
	}

	// Handle stdin and messages
	info, _ := os.Stdin.Stat()
	pipedToStdin := (info.Mode() & os.ModeCharDevice) == 0

	// Append positional arguments to the message (custom message)
	if len(args) > 0 {
		ret.Message = AppendMessage(ret.Message, args[len(args)-1])
	}

	if pipedToStdin {
		var pipedMessage string
		if pipedMessage, err = readStdin(); err != nil {
			return
		}
		ret.Message = AppendMessage(ret.Message, pipedMessage)
	}

	return ret, nil
}

func assignWithConversion(targetField, sourceField reflect.Value) error {
	// Handle string source values
	if sourceField.Kind() == reflect.String {
		str := sourceField.String()
		switch targetField.Kind() {
		case reflect.Int:
			// Try parsing as float first to handle "42.9" -> 42
			if val, err := strconv.ParseFloat(str, 64); err == nil {
				targetField.SetInt(int64(val))
				return nil
			}
			// Try direct int parse
			if val, err := strconv.ParseInt(str, 10, 64); err == nil {
				targetField.SetInt(val)
				return nil
			}
		case reflect.Float64:
			if val, err := strconv.ParseFloat(str, 64); err == nil {
				targetField.SetFloat(val)
				return nil
			}
		case reflect.Bool:
			if val, err := strconv.ParseBool(str); err == nil {
				targetField.SetBool(val)
				return nil
			}
		}
		return fmt.Errorf("cannot convert string %q to %v", str, targetField.Kind())
	}

	return fmt.Errorf("unsupported conversion from %v to %v", sourceField.Kind(), targetField.Kind())
}

func loadYAMLConfig(configPath string) (*Flags, error) {
	absPath, err := common.GetAbsolutePath(configPath)
	if err != nil {
		return nil, fmt.Errorf("invalid config path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", absPath)
		}
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Use the existing Flags struct for YAML unmarshal
	config := &Flags{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	Debugf("Config: %v\n", config)

	return config, nil
}

// readStdin reads from stdin and returns the input as a string or an error
func readStdin() (ret string, err error) {
	reader := bufio.NewReader(os.Stdin)
	var sb strings.Builder
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				sb.WriteString(line)
				break
			}
			return "", fmt.Errorf("error reading piped message from stdin: %w", err)
		}
		sb.WriteString(line)
	}
	return sb.String(), nil
}

func (o *Flags) BuildChatOptions() (ret *common.ChatOptions) {
	ret = &common.ChatOptions{
		Temperature:        o.Temperature,
		TopP:               o.TopP,
		PresencePenalty:    o.PresencePenalty,
		FrequencyPenalty:   o.FrequencyPenalty,
		Raw:                o.Raw,
		Seed:               o.Seed,
		ModelContextLength: o.ModelContextLength,
	}
	return
}

func (o *Flags) BuildChatRequest(Meta string) (ret *common.ChatRequest, err error) {
	ret = &common.ChatRequest{
		ContextName:      o.Context,
		SessionName:      o.Session,
		PatternName:      o.Pattern,
		PatternVariables: o.PatternVariables,
		InputHasVars:     o.InputHasVars,
		Meta:             Meta,
	}

	var message *goopenai.ChatCompletionMessage
	if o.Attachments == nil || len(o.Attachments) == 0 {
		if o.Message != "" {
			message = &goopenai.ChatCompletionMessage{
				Role:    goopenai.ChatMessageRoleUser,
				Content: strings.TrimSpace(o.Message),
			}
		}
	} else {
		message = &goopenai.ChatCompletionMessage{
			Role: goopenai.ChatMessageRoleUser,
		}

		if o.Message != "" {
			message.MultiContent = append(message.MultiContent, goopenai.ChatMessagePart{
				Type: goopenai.ChatMessagePartTypeText,
				Text: strings.TrimSpace(o.Message),
			})
		}

		for _, attachmentValue := range o.Attachments {
			var attachment *common.Attachment
			if attachment, err = common.NewAttachment(attachmentValue); err != nil {
				return
			}
			url := attachment.URL
			if url == nil {
				var base64Image string
				if base64Image, err = attachment.Base64Content(); err != nil {
					return
				}
				var mimeType string
				if mimeType, err = attachment.ResolveType(); err != nil {
					return
				}
				dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image)
				url = &dataURL
			}
			message.MultiContent = append(message.MultiContent, goopenai.ChatMessagePart{
				Type: goopenai.ChatMessagePartTypeImageURL,
				ImageURL: &goopenai.ChatMessageImageURL{
					URL: *url,
				},
			})
		}
	}
	ret.Message = message

	if o.Language != "" {
		if langTag, langErr := language.Parse(o.Language); langErr == nil {
			ret.Language = langTag.String()
		}
	}
	return
}

func (o *Flags) AppendMessage(message string) {
	o.Message = AppendMessage(o.Message, message)
	return
}

func (o *Flags) IsChatRequest() (ret bool) {
	ret = o.Message != "" || len(o.Attachments) > 0 || o.Context != "" || o.Session != "" || o.Pattern != ""
	return
}

func (o *Flags) WriteOutput(message string) (err error) {
	fmt.Println(message)
	if o.Output != "" {
		err = CreateOutputFile(message, o.Output)
	}
	return
}

func AppendMessage(message string, newMessage string) (ret string) {
	if message != "" {
		ret = message + "\n" + newMessage
	} else {
		ret = newMessage
	}
	return
}
