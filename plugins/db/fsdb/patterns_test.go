package fsdb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestPatternsEntity(t *testing.T) (*PatternsEntity, func()) {
	// Create a temporary directory for test patterns
	tmpDir, err := os.MkdirTemp("", "test-patterns-*")
	require.NoError(t, err)

	entity := &PatternsEntity{
		StorageEntity: &StorageEntity{
			Dir:       tmpDir,
			Label:     "patterns",
			ItemIsDir: true,
		},
		SystemPatternFile: "system.md",
	}

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return entity, cleanup
}

// Helper to create a test pattern file
func createTestPattern(t *testing.T, entity *PatternsEntity, name, content string) {
	patternDir := filepath.Join(entity.Dir, name)
	err := os.MkdirAll(patternDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(patternDir, entity.SystemPatternFile), []byte(content), 0644)
	require.NoError(t, err)
}

func TestApplyVariables(t *testing.T) {
	entity := &PatternsEntity{}

	tests := []struct {
		name      string
		pattern   *Pattern
		variables map[string]string
		input     string
		want      string
		wantErr   bool
	}{
		{
			name: "pattern with explicit input placement",
			pattern: &Pattern{
				Pattern: "You are a {{role}}.\n{{input}}\nPlease analyze.",
			},
			variables: map[string]string{
				"role": "security expert",
			},
			input: "Check this code",
			want:  "You are a security expert.\nCheck this code\nPlease analyze.",
		},
		{
			name: "pattern without input variable gets input appended",
			pattern: &Pattern{
				Pattern: "You are a {{role}}.\nPlease analyze.",
			},
			variables: map[string]string{
				"role": "code reviewer",
			},
			input: "Review this PR",
			want:  "You are a code reviewer.\nPlease analyze.\nReview this PR",
		},
		// ... previous test cases ...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := entity.applyVariables(tt.pattern, tt.variables, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, tt.pattern.Pattern)
		})
	}
}

func TestGetApplyVariables(t *testing.T) {
	entity, cleanup := setupTestPatternsEntity(t)
	defer cleanup()

	// Create a test pattern
	createTestPattern(t, entity, "test-pattern", "You are a {{role}}.\n{{input}}")

	tests := []struct {
		name      string
		source    string
		variables map[string]string
		input     string
		want      string
		wantErr   bool
	}{
		{
			name:   "basic pattern with variables and input",
			source: "test-pattern",
			variables: map[string]string{
				"role": "reviewer",
			},
			input: "check this code",
			want:  "You are a reviewer.\ncheck this code",
		},
		{
			name:      "pattern with missing variable",
			source:    "test-pattern",
			variables: map[string]string{},
			input:     "test input",
			wantErr:   true,
		},
		{
			name:    "non-existent pattern",
			source:  "non-existent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := entity.GetApplyVariables(tt.source, tt.variables, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, result.Pattern)
		})
	}
}
