package sse

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCollectStreamDedupesContinueSnapshotReplay(t *testing.T) {
	body := strings.Join([]string{
		`data: {"v":{"response":{"fragments":[{"id":2,"type":"THINK","content":"我们","references":[],"stage_id":1}]}}}`,
		``,
		`data: {"p":"response/fragments/-1/content","o":"APPEND","v":"被"}`,
		``,
		`data: {"v":"问到"}`,
		``,
		`data: {"p":"response/status","v":"INCOMPLETE"}`,
		``,
		`data: {"v":{"response":{"fragments":[{"id":2,"type":"THINK","content":"我们被问到继续","references":[],"stage_id":1}]}}}`,
		``,
		`data: {"v":"分析"}`,
		``,
		`data: {"p":"response/status","v":"FINISHED"}`,
		``,
	}, "\n")

	resp := &http.Response{Body: io.NopCloser(strings.NewReader(body))}
	got := CollectStream(resp, true, true)
	if got.Thinking != "我们被问到继续分析" {
		t.Fatalf("unexpected thinking after dedupe: %q", got.Thinking)
	}
}
