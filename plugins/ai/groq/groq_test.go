package groq

// Test generated using Keploy
import (
	"testing"
)

func TestNewClientEmbeddedClientNotNil(t *testing.T) {
	client := NewClient()
	if client.Client == nil {
		t.Fatalf("Expected embedded openai.Client to be non-nil, got nil")
	}
}
