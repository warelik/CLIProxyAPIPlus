package claude

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func convertGeminiCLIChunks(t *testing.T, chunks ...[]byte) string {
	t.Helper()
	requestJSON := []byte(`{"model":"gemini-test","messages":[{"role":"user","content":[{"type":"text","text":"hi"}]}]}`)
	var param any
	ctx := context.Background()
	output := make([]byte, 0, 1024)
	for _, chunk := range chunks {
		output = append(output, bytes.Join(ConvertGeminiCLIResponseToClaude(ctx, "gemini-test", requestJSON, requestJSON, chunk, &param), nil)...)
	}
	output = append(output, bytes.Join(ConvertGeminiCLIResponseToClaude(ctx, "gemini-test", requestJSON, requestJSON, []byte("[DONE]"), &param), nil)...)
	return string(output)
}

func TestConvertGeminiCLIResponseToClaude_SignatureOnlyPartEmitsThinkingSignature(t *testing.T) {
	thinkingChunk := []byte(`{"response":{"candidates":[{"content":{"parts":[{"text":"thinking text","thought":true}]}}],"modelVersion":"gemini-test","responseId":"resp-test"}}`)
	signatureChunk := []byte(`{"response":{"candidates":[{"content":{"parts":[{"text":"","thoughtSignature":"sig-test"}]},"finishReason":"STOP"}],"modelVersion":"gemini-test","responseId":"resp-test"}}`)

	outputText := convertGeminiCLIChunks(t, thinkingChunk, signatureChunk)

	if !strings.Contains(outputText, `"type":"signature_delta"`) || !strings.Contains(outputText, `"signature":"sig-test"`) {
		t.Fatalf("signature-only part must be emitted as a thinking signature delta: %s", outputText)
	}
	if strings.Contains(outputText, `"content_block":{"type":"text"`) {
		t.Fatalf("signature-only part must not open an empty text block: %s", outputText)
	}
	if got := strings.Count(outputText, `"type":"content_block_start"`); got != 1 {
		t.Fatalf("expected the signature to reuse the open thinking block, got %d content_block_start events: %s", got, outputText)
	}
}

func TestConvertGeminiCLIResponseToClaude_ThoughtPartCarriesSignature(t *testing.T) {
	chunk := []byte(`{"response":{"candidates":[{"content":{"parts":[{"text":"reasoning","thought":true,"thoughtSignature":"sig-thought"}]}}],"modelVersion":"gemini-test","responseId":"resp-test"}}`)

	outputText := convertGeminiCLIChunks(t, chunk)

	if !strings.Contains(outputText, `"type":"thinking_delta"`) || !strings.Contains(outputText, `"thinking":"reasoning"`) {
		t.Fatalf("thought text must still be emitted as a thinking delta: %s", outputText)
	}
	if !strings.Contains(outputText, `"signature":"sig-thought"`) {
		t.Fatalf("thought signature must be emitted alongside the thinking block: %s", outputText)
	}
	if strings.Index(outputText, `"thinking":"reasoning"`) > strings.Index(outputText, `"signature":"sig-thought"`) {
		t.Fatalf("thinking delta must precede its signature delta: %s", outputText)
	}
}

func TestConvertGeminiCLIResponseToClaude_VisibleTextWithSignatureStaysVisible(t *testing.T) {
	chunk := []byte(`{"response":{"candidates":[{"content":{"parts":[{"text":"visible answer","thoughtSignature":"sig-visible"}]}}],"modelVersion":"gemini-test","responseId":"resp-test"}}`)

	outputText := convertGeminiCLIChunks(t, chunk)

	if !strings.Contains(outputText, `"type":"text_delta"`) || !strings.Contains(outputText, `"text":"visible answer"`) {
		t.Fatalf("visible text must stay a text delta: %s", outputText)
	}
	if strings.Contains(outputText, `"thinking":"visible answer"`) {
		t.Fatalf("visible text must never be rerouted into a thinking block: %s", outputText)
	}
	if !strings.Contains(outputText, `"signature":"sig-visible"`) {
		t.Fatalf("signature attached to a visible text part must still be forwarded: %s", outputText)
	}
}

func TestConvertGeminiCLIResponseToClaude_SnakeCaseThoughtSignatureIsAccepted(t *testing.T) {
	chunk := []byte(`{"response":{"candidates":[{"content":{"parts":[{"text":"reasoning","thought":true,"thought_signature":"sig-snake"}]}}],"modelVersion":"gemini-test","responseId":"resp-test"}}`)

	outputText := convertGeminiCLIChunks(t, chunk)

	if !strings.Contains(outputText, `"signature":"sig-snake"`) {
		t.Fatalf("snake_case thought_signature must be accepted: %s", outputText)
	}
}

func TestConvertGeminiCLIResponseToClaude_WithoutSignatureEmitsNoSignatureDelta(t *testing.T) {
	chunk := []byte(`{"response":{"candidates":[{"content":{"parts":[{"text":"reasoning","thought":true},{"text":"answer"}]}}],"modelVersion":"gemini-test","responseId":"resp-test"}}`)

	outputText := convertGeminiCLIChunks(t, chunk)

	if strings.Contains(outputText, `"type":"signature_delta"`) {
		t.Fatalf("no signature in the payload must not synthesize a signature delta: %s", outputText)
	}
	if !strings.Contains(outputText, `"thinking":"reasoning"`) || !strings.Contains(outputText, `"text":"answer"`) {
		t.Fatalf("unsigned parts must keep their original routing: %s", outputText)
	}
}
