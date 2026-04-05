package sse

import "strings"

// TrimContinuationOverlap removes the already-seen prefix when DeepSeek
// continue rounds resend the full fragment snapshot instead of only the new
// suffix. Non-overlapping chunks are returned unchanged.
func TrimContinuationOverlap(existing, incoming string) string {
	if incoming == "" {
		return ""
	}
	if existing == "" {
		return incoming
	}
	if strings.HasPrefix(incoming, existing) {
		return incoming[len(existing):]
	}
	if strings.HasPrefix(existing, incoming) {
		return ""
	}
	return incoming
}
