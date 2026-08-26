package auth

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestSummarizeErrorForLogRedactsSecrets(t *testing.T) {
	got := summarizeErrorForLog(errors.New("Incorrect API key provided: sk-live-secret"))
	if strings.Contains(got, "sk-live-secret") {
		t.Fatalf("summarizeErrorForLog leaked API key: %q", got)
	}
	if !strings.Contains(got, "[REDACTED]") {
		t.Fatalf("summarizeErrorForLog = %q, want [REDACTED]", got)
	}

	got = summarizeErrorForLog(errors.New("authorization failed: Bearer abcdefghijklmnop"))
	if strings.Contains(got, "abcdefghijklmnop") {
		t.Fatalf("summarizeErrorForLog leaked bearer token: %q", got)
	}
}

func TestRedactSecretsForLog_QuotedJSON(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		leaks []string
	}{
		{
			name:  "apiKey with non-sk secret",
			in:    `{"apiKey":"AIza-secret"}`,
			leaks: []string{"AIza-secret"},
		},
		{
			name:  "token key",
			in:    `{"token":"foo"}`,
			leaks: []string{"foo"},
		},
		{
			name:  "authorization key with quoted value",
			in:    `{"authorization":"AIza-secret"}`,
			leaks: []string{"AIza-secret"},
		},
		{
			name:  "api-key with unquoted key",
			in:    `api-key:AIza-secret`,
			leaks: []string{"AIza-secret"},
		},
		{
			name:  "sk prefix still redacted",
			in:    `sk-live-secret`,
			leaks: []string{"sk-live-secret"},
		},
		{
			name:  "prose API key with xAI prefix",
			in:    "Incorrect API key provided: xai-live-secret",
			leaks: []string{"xai-live-secret"},
		},
		{
			name:  "prose API key with unlabeled value",
			in:    "Incorrect API key provided: naked-secret-value12",
			leaks: []string{"naked-secret-value12"},
		},
		{
			name:  "groq prefix without label",
			in:    "invalid key gsk_live-secret-value",
			leaks: []string{"gsk_live-secret-value"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := redactSecretsForLog(tc.in)
			for _, leak := range tc.leaks {
				if strings.Contains(got, leak) {
					t.Fatalf("redacted = %q, contains %q", got, leak)
				}
			}
			if !strings.Contains(got, "[REDACTED]") {
				t.Fatalf("redacted = %q, want [REDACTED]", got)
			}
		})
	}
}

func TestResultErrorFromErrorRedactsSecrets(t *testing.T) {
	leaky := &Error{Code: "sk-xyz1234567890", Message: "Bearer sk-abc123def456", HTTPStatus: http.StatusUnauthorized}
	result := resultErrorFromError(leaky)
	if result == nil {
		t.Fatal("resultErrorFromError returned nil")
	}
	if strings.Contains(result.Message, "sk-abc123def456") {
		t.Fatalf("result error message = %q, want secret redacted", result.Message)
	}
	if !strings.Contains(result.Message, "REDACTED") {
		t.Fatalf("result error message = %q, want [REDACTED] placeholder", result.Message)
	}
	if strings.Contains(result.Code, "sk-xyz1234567890") {
		t.Fatalf("result error code = %q, want secret redacted", result.Code)
	}
	if !strings.Contains(result.Code, "REDACTED") {
		t.Fatalf("result error code = %q, want [REDACTED] placeholder", result.Code)
	}
}
