package openai

import (
	"net/http"
	"strings"

	"ds2api/internal/sse"
)

func writeUpstreamEmptyOutputError(w http.ResponseWriter, result sse.CollectResult) bool {
	if strings.TrimSpace(result.Thinking) != "" || strings.TrimSpace(sanitizeLeakedOutput(result.Text)) != "" {
		return false
	}
	if result.ContentFilter {
		writeOpenAIErrorWithCode(w, http.StatusBadRequest, "Upstream content filtered the response and returned no output.", "content_filter")
		return true
	}
	writeOpenAIErrorWithCode(w, http.StatusBadGateway, "Upstream model returned empty output.", "upstream_empty_output")
	return true
}
