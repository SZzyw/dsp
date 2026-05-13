package sse

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
)

const (
	obfuscationTargetLen = 512
	obfuscationMinPad    = 16
)

func AddObfuscation(chunk map[string]any) {
	choices, ok := chunk["choices"].([]any)
	if !ok || len(choices) == 0 {
		return
	}
	choice, ok := choices[0].(map[string]any)
	if !ok {
		return
	}
	delta, ok := choice["delta"].(map[string]any)
	if !ok {
		return
	}

	_, hasContent := delta["content"]
	_, hasReasoning := delta["reasoning_content"]
	_, hasFinish := choice["finish_reason"]
	if !hasContent && !hasReasoning && !hasFinish {
		return
	}

	without, _ := json.Marshal(chunk)
	needed := obfuscationTargetLen - len(without) - len(`,"obfuscation":""`)
	if needed < obfuscationMinPad {
		needed = obfuscationMinPad
	}

	pad := make([]byte, (needed*3)/4)
	rand.Read(pad)
	delta["obfuscation"] = base64.StdEncoding.EncodeToString(pad)[:needed]
}
