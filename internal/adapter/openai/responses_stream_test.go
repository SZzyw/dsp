package openai

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleResponsesStreamNoDuplicateTailInCompletedOutputText(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	rec := httptest.NewRecorder()

	sseLine := func(v string) string {
		b, _ := json.Marshal(map[string]any{
			"p": "response/content",
			"v": v,
		})
		return "data: " + string(b) + "\n"
	}

	tail := `{"tool_calls":[{"name":"read_file","input":`
	streamBody := sseLine("Before ") + sseLine(tail) + "data: [DONE]\n"
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(streamBody)),
	}

	h.handleResponsesStream(rec, req, resp, "owner-a", "resp_test", "deepseek-chat", "prompt", false, false, []string{"read_file"})

	completed, ok := extractSSEEventPayload(rec.Body.String(), "response.completed")
	if !ok {
		t.Fatalf("expected response.completed event, body=%s", rec.Body.String())
	}
	responseObj, _ := completed["response"].(map[string]any)
	outputText, _ := responseObj["output_text"].(string)
	if strings.Count(outputText, tail) != 1 {
		t.Fatalf("expected tail to appear once in output_text, got output_text=%q", outputText)
	}
}

func extractSSEEventPayload(body, targetEvent string) (map[string]any, bool) {
	scanner := bufio.NewScanner(strings.NewReader(body))
	matched := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "event: ") {
			evt := strings.TrimSpace(strings.TrimPrefix(line, "event: "))
			matched = evt == targetEvent
			continue
		}
		if !matched || !strings.HasPrefix(line, "data: ") {
			continue
		}
		raw := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
		if raw == "" || raw == "[DONE]" {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return nil, false
		}
		return payload, true
	}
	return nil, false
}
