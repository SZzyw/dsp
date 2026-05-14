package openai

import (
	"strings"
	"testing"

	"ds2api/internal/config"
	"ds2api/internal/promptcompat"
)

type mockOpenAIConfig struct {
	aliases            map[string]string
	autoDeleteMode     string
	toolMode           string
	earlyEmit          string
	responsesTTL       int
	embedProv          string
	currentInputFlash  *bool
	currentInputPro    *bool
	currentInputVision *bool
	familyPolicy       *config.ModelFamilyPolicyConfig
	toolCallsEnabled   *bool
}

func (m mockOpenAIConfig) ModelAliases() map[string]string     { return m.aliases }
func (m mockOpenAIConfig) ToolcallMode() string                { return m.toolMode }
func (m mockOpenAIConfig) ToolcallEarlyEmitConfidence() string { return m.earlyEmit }
func (m mockOpenAIConfig) ResponsesStoreTTLSeconds() int       { return m.responsesTTL }
func (m mockOpenAIConfig) EmbeddingsProvider() string          { return m.embedProv }
func (m mockOpenAIConfig) AutoDeleteMode() string {
	if m.autoDeleteMode == "" {
		return "none"
	}
	return m.autoDeleteMode
}
func currentInputEnabled(ptr *bool) bool {
	if ptr == nil {
		return false
	}
	return *ptr
}
func (m mockOpenAIConfig) AutoDeleteSessions() bool { return false }
func (m mockOpenAIConfig) CurrentInputFileEnabledForModel(model string) bool {
	switch model {
	case "deepseek-v4-flash", "deepseek-v4-flash-search", "deepseek-v4-flash-nothinking":
		return currentInputEnabled(m.currentInputFlash)
	case "deepseek-v4-pro", "deepseek-v4-pro-search", "deepseek-v4-pro-nothinking":
		return currentInputEnabled(m.currentInputPro)
	case "deepseek-v4-vision", "deepseek-v4-vision-nothinking":
		return currentInputEnabled(m.currentInputVision)
	default:
		return true
	}
}
func (m mockOpenAIConfig) ModelFamilyPolicy() config.ModelFamilyPolicyConfig {
	if m.familyPolicy == nil {
		return config.ModelFamilyPolicyConfig{}
	}
	return *m.familyPolicy
}
func (m mockOpenAIConfig) ToolCallsEnabledForModel(model string) bool {
	if m.toolCallsEnabled == nil {
		return true
	}
	return *m.toolCallsEnabled
}

func TestNormalizeOpenAIChatRequestWithConfigInterface(t *testing.T) {
	cfg := mockOpenAIConfig{
		aliases: map[string]string{
			"my-model": "deepseek-v4-flash-search",
		},
	}
	req := map[string]any{
		"model":    "my-model",
		"messages": []any{map[string]any{"role": "user", "content": "hello"}},
	}
	out, err := promptcompat.NormalizeOpenAIChatRequest(cfg, req, "")
	if err != nil {
		t.Fatalf("promptcompat.NormalizeOpenAIChatRequest error: %v", err)
	}
	if out.ResolvedModel != "deepseek-v4-flash-search" {
		t.Fatalf("resolved model mismatch: got=%q", out.ResolvedModel)
	}
	if !out.Search || !out.Thinking {
		t.Fatalf("unexpected model flags: thinking=%v search=%v", out.Thinking, out.Search)
	}
}

func TestNormalizeOpenAIChatRequestDisablesThinkingForNoThinkingModel(t *testing.T) {
	cfg := mockOpenAIConfig{}
	req := map[string]any{
		"model":            "deepseek-v4-pro-nothinking",
		"messages":         []any{map[string]any{"role": "user", "content": "hello"}},
		"reasoning_effort": "high",
	}
	out, err := promptcompat.NormalizeOpenAIChatRequest(cfg, req, "")
	if err != nil {
		t.Fatalf("promptcompat.NormalizeOpenAIChatRequest error: %v", err)
	}
	if out.ResolvedModel != "deepseek-v4-pro-nothinking" {
		t.Fatalf("resolved model mismatch: got=%q", out.ResolvedModel)
	}
	if out.Thinking {
		t.Fatalf("expected nothinking model to force thinking off")
	}
	if out.Search {
		t.Fatalf("expected search=false for deepseek-v4-pro-nothinking, got=%v", out.Search)
	}
}

func TestNormalizeOpenAIChatRequestRoutesModelFamily(t *testing.T) {
	cfg := mockOpenAIConfig{
		familyPolicy: &config.ModelFamilyPolicyConfig{
			Pro: config.ModelFamilyPolicyRule{Mode: "route", Target: "flash"},
		},
	}
	req := map[string]any{
		"model":    "deepseek-v4-pro",
		"messages": []any{map[string]any{"role": "user", "content": "hello"}},
	}
	out, err := promptcompat.NormalizeOpenAIChatRequest(cfg, req, "")
	if err != nil {
		t.Fatalf("promptcompat.NormalizeOpenAIChatRequest error: %v", err)
	}
	if out.ResolvedModel != "deepseek-v4-flash" {
		t.Fatalf("expected routed model deepseek-v4-flash, got %q", out.ResolvedModel)
	}
	if out.ResponseModel != "deepseek-v4-pro" {
		t.Fatalf("expected response model preserved, got %q", out.ResponseModel)
	}
}

func TestNormalizeOpenAIResponsesRequestAlwaysAcceptsWideInput(t *testing.T) {
	req := map[string]any{
		"model": "deepseek-v4-flash",
		"input": "hi",
	}

	out, err := promptcompat.NormalizeOpenAIResponsesRequest(mockOpenAIConfig{
		aliases: map[string]string{},
	}, req, "")
	if err != nil {
		t.Fatalf("unexpected error for wide input request: %v", err)
	}
	if out.Surface != "openai_responses" {
		t.Fatalf("unexpected surface: %q", out.Surface)
	}
	if !strings.Contains(out.FinalPrompt, "<|User|>hi") {
		t.Fatalf("unexpected final prompt: %q", out.FinalPrompt)
	}
}

func TestNormalizeOpenAIChatRequestDropsToolsWhenToolPolicyDisabled(t *testing.T) {
	disabled := false
	req := map[string]any{
		"model": "deepseek-v4-flash",
		"messages": []any{
			map[string]any{"role": "user", "content": "hello"},
		},
		"tools": []any{
			map[string]any{
				"type":     "function",
				"function": map[string]any{"name": "search"},
			},
		},
	}

	out, err := promptcompat.NormalizeOpenAIChatRequest(mockOpenAIConfig{toolCallsEnabled: &disabled}, req, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ToolCallsEnabled {
		t.Fatal("expected tool calls to be disabled")
	}
	if out.ToolsRaw == nil {
		t.Fatal("expected original tools preserved for suppression")
	}
	if len(out.ToolNames) == 0 {
		t.Fatal("expected tool names preserved for suppression")
	}
}

func TestNormalizeOpenAIResponsesRequestDropsToolsWhenToolPolicyDisabled(t *testing.T) {
	disabled := false
	req := map[string]any{
		"model": "deepseek-v4-flash",
		"input": "hi",
		"tools": []any{
			map[string]any{
				"type":     "function",
				"function": map[string]any{"name": "search"},
			},
		},
	}

	out, err := promptcompat.NormalizeOpenAIResponsesRequest(mockOpenAIConfig{toolCallsEnabled: &disabled}, req, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ToolCallsEnabled {
		t.Fatal("expected tool calls to be disabled")
	}
	if out.ToolsRaw == nil {
		t.Fatal("expected original tools preserved for suppression")
	}
	if len(out.ToolNames) == 0 {
		t.Fatal("expected tool names preserved for suppression")
	}
	if out.ToolChoice.Mode != promptcompat.ToolChoiceNone {
		t.Fatalf("expected tool_choice none, got %#v", out.ToolChoice)
	}
}
