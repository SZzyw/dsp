package config

import (
	"fmt"
	"strings"
	"time"
)

type ModelInfo struct {
	ID         string `json:"id"`
	Object     string `json:"object"`
	Created    int64  `json:"created"`
	OwnedBy    string `json:"owned_by"`
	Permission []any  `json:"permission,omitempty"`
}
type OllamaModelInfo struct {
	Name       string `json:"name"`
	Model      string `json:"model"`
	Size       int64  `json:"size"`
	ModifiedAt string `json:"modified_at"`
}
type OllamaCapabilitiesModelInfo struct {
	ID           string   `json:"id"`
	Capabilities []string `json:"capabilities"`
}

type ModelAliasReader interface {
	ModelAliases() map[string]string
}

type ModelFamilyPolicyReader interface {
	ModelAliases() map[string]string
	ModelFamilyPolicy() ModelFamilyPolicyConfig
}

const noThinkingModelSuffix = "-nothinking"

var deepSeekBaseModels = []ModelInfo{
	{ID: "deepseek-v4-flash", Object: "model", Created: 1677610602, OwnedBy: "deepseek", Permission: []any{}},
	{ID: "deepseek-v4-pro", Object: "model", Created: 1677610602, OwnedBy: "deepseek", Permission: []any{}},
	{ID: "deepseek-v4-flash-search", Object: "model", Created: 1677610602, OwnedBy: "deepseek", Permission: []any{}},
	{ID: "deepseek-v4-pro-search", Object: "model", Created: 1677610602, OwnedBy: "deepseek", Permission: []any{}},
	{ID: "deepseek-v4-vision", Object: "model", Created: 1677610602, OwnedBy: "deepseek", Permission: []any{}},
}

var OllamaCapabilitiesModels = []OllamaCapabilitiesModelInfo{
	{ID: "deepseek-v4-flash", Capabilities: []string{"tools", "thinking"}},
	{ID: "deepseek-v4-pro", Capabilities: []string{"tools", "thinking"}},
	{ID: "deepseek-v4-flash-search", Capabilities: []string{"tools", "thinking"}},
	{ID: "deepseek-v4-pro-search", Capabilities: []string{"tools", "thinking"}},
	{ID: "deepseek-v4-vision", Capabilities: []string{"tools", "thinking", "vision"}},
	{ID: "deepseek-v4-flash-nothinking", Capabilities: []string{"tools"}},
	{ID: "deepseek-v4-pro-nothinking", Capabilities: []string{"tools"}},
	{ID: "deepseek-v4-flash-search-nothinking", Capabilities: []string{"tools"}},
	{ID: "deepseek-v4-pro-search-nothinking", Capabilities: []string{"tools"}},
	{ID: "deepseek-v4-vision-nothinking", Capabilities: []string{"tools", "vision"}},
}

var DeepSeekModels = appendNoThinkingVariants(deepSeekBaseModels)
var OllamaModels = mapToOllamaModels(DeepSeekModels)
var claudeBaseModels = []ModelInfo{
	// Current aliases
	{ID: "claude-opus-4-6", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-sonnet-4-6", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-haiku-4-5", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},

	// Claude 4.x snapshots and prior aliases kept for compatibility
	{ID: "claude-sonnet-4-5", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-opus-4-1", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-opus-4-1-20250805", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-opus-4-0", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-opus-4-20250514", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-sonnet-4-5-20250929", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-sonnet-4-0", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-sonnet-4-20250514", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-haiku-4-5-20251001", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},

	// Claude 3.x (legacy/deprecated snapshots and aliases)
	{ID: "claude-3-7-sonnet-latest", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-3-7-sonnet-20250219", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-3-5-sonnet-latest", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-3-5-sonnet-20240620", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-3-5-sonnet-20241022", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-3-opus-20240229", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-3-sonnet-20240229", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-3-5-haiku-latest", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-3-5-haiku-20241022", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-3-haiku-20240307", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
}

var ClaudeModels = appendNoThinkingVariants(claudeBaseModels)

func GetModelConfig(model string) (thinking bool, search bool, ok bool) {
	baseModel, noThinking := splitNoThinkingModel(model)
	if baseModel == "" {
		return false, false, false
	}
	switch baseModel {
	case "deepseek-v4-flash", "deepseek-v4-pro", "deepseek-v4-vision":
		return !noThinking, false, true
	case "deepseek-v4-flash-search", "deepseek-v4-pro-search":
		return !noThinking, true, true
	default:
		return false, false, false
	}
}

func GetModelType(model string) (modelType string, ok bool) {
	baseModel, _ := splitNoThinkingModel(model)
	switch baseModel {
	case "deepseek-v4-flash", "deepseek-v4-flash-search":
		return "default", true
	case "deepseek-v4-pro", "deepseek-v4-pro-search":
		return "expert", true
	case "deepseek-v4-vision":
		return "vision", true
	default:
		return "", false
	}
}

func IsSupportedDeepSeekModel(model string) bool {
	_, _, ok := GetModelConfig(model)
	return ok
}

func IsNoThinkingModel(model string) bool {
	_, noThinking := splitNoThinkingModel(model)
	return noThinking
}

func DefaultModelAliases() map[string]string {
	return map[string]string{
		// OpenAI GPT / ChatGPT families
		"chatgpt-4o":          "deepseek-v4-flash",
		"gpt-4":               "deepseek-v4-flash",
		"gpt-4-turbo":         "deepseek-v4-flash",
		"gpt-4-turbo-preview": "deepseek-v4-flash",
		"gpt-4.5-preview":     "deepseek-v4-flash",
		"gpt-4o":              "deepseek-v4-flash",
		"gpt-4o-mini":         "deepseek-v4-flash",
		"gpt-4.1":             "deepseek-v4-flash",
		"gpt-4.1-mini":        "deepseek-v4-flash",
		"gpt-4.1-nano":        "deepseek-v4-flash",
		"gpt-5":               "deepseek-v4-flash",
		"gpt-5-chat":          "deepseek-v4-flash",
		"gpt-5.1":             "deepseek-v4-flash",
		"gpt-5.1-chat":        "deepseek-v4-flash",
		"gpt-5.2":             "deepseek-v4-flash",
		"gpt-5.2-chat":        "deepseek-v4-flash",
		"gpt-5.3-chat":        "deepseek-v4-flash",
		"gpt-5.4":             "deepseek-v4-flash",
		"gpt-5.5":             "deepseek-v4-flash",
		"gpt-5-mini":          "deepseek-v4-flash",
		"gpt-5-nano":          "deepseek-v4-flash",
		"gpt-5.4-mini":        "deepseek-v4-flash",
		"gpt-5.4-nano":        "deepseek-v4-flash",
		"gpt-5-pro":           "deepseek-v4-pro",
		"gpt-5.2-pro":         "deepseek-v4-pro",
		"gpt-5.4-pro":         "deepseek-v4-pro",
		"gpt-5.5-pro":         "deepseek-v4-pro",
		"gpt-5-codex":         "deepseek-v4-pro",
		"gpt-5.1-codex":       "deepseek-v4-pro",
		"gpt-5.1-codex-mini":  "deepseek-v4-pro",
		"gpt-5.1-codex-max":   "deepseek-v4-pro",
		"gpt-5.2-codex":       "deepseek-v4-pro",
		"gpt-5.3-codex":       "deepseek-v4-pro",
		"codex-mini-latest":   "deepseek-v4-pro",

		// OpenAI reasoning / research families
		"o1":                    "deepseek-v4-pro",
		"o1-preview":            "deepseek-v4-pro",
		"o1-mini":               "deepseek-v4-pro",
		"o1-pro":                "deepseek-v4-pro",
		"o3":                    "deepseek-v4-pro",
		"o3-mini":               "deepseek-v4-pro",
		"o3-pro":                "deepseek-v4-pro",
		"o3-deep-research":      "deepseek-v4-pro-search",
		"o4-mini":               "deepseek-v4-pro",
		"o4-mini-deep-research": "deepseek-v4-pro-search",

		// Claude current and historical aliases
		"claude-opus-4-6":            "deepseek-v4-pro",
		"claude-opus-4-1":            "deepseek-v4-pro",
		"claude-opus-4-1-20250805":   "deepseek-v4-pro",
		"claude-opus-4-0":            "deepseek-v4-pro",
		"claude-opus-4-20250514":     "deepseek-v4-pro",
		"claude-sonnet-4-6":          "deepseek-v4-flash",
		"claude-sonnet-4-5":          "deepseek-v4-flash",
		"claude-sonnet-4-5-20250929": "deepseek-v4-flash",
		"claude-sonnet-4-0":          "deepseek-v4-flash",
		"claude-sonnet-4-20250514":   "deepseek-v4-flash",
		"claude-haiku-4-5":           "deepseek-v4-flash",
		"claude-haiku-4-5-20251001":  "deepseek-v4-flash",
		"claude-3-7-sonnet":          "deepseek-v4-flash",
		"claude-3-7-sonnet-latest":   "deepseek-v4-flash",
		"claude-3-7-sonnet-20250219": "deepseek-v4-flash",
		"claude-3-5-sonnet":          "deepseek-v4-flash",
		"claude-3-5-sonnet-latest":   "deepseek-v4-flash",
		"claude-3-5-sonnet-20240620": "deepseek-v4-flash",
		"claude-3-5-sonnet-20241022": "deepseek-v4-flash",
		"claude-3-5-haiku":           "deepseek-v4-flash",
		"claude-3-5-haiku-latest":    "deepseek-v4-flash",
		"claude-3-5-haiku-20241022":  "deepseek-v4-flash",
		"claude-3-opus":              "deepseek-v4-pro",
		"claude-3-opus-20240229":     "deepseek-v4-pro",
		"claude-3-sonnet":            "deepseek-v4-flash",
		"claude-3-sonnet-20240229":   "deepseek-v4-flash",
		"claude-3-haiku":             "deepseek-v4-flash",
		"claude-3-haiku-20240307":    "deepseek-v4-flash",

		// Gemini current and historical text / multimodal models
		"gemini-pro":            "deepseek-v4-pro",
		"gemini-pro-vision":     "deepseek-v4-vision",
		"gemini-pro-latest":     "deepseek-v4-pro",
		"gemini-flash-latest":   "deepseek-v4-flash",
		"gemini-1.5-pro":        "deepseek-v4-pro",
		"gemini-1.5-flash":      "deepseek-v4-flash",
		"gemini-1.5-flash-8b":   "deepseek-v4-flash",
		"gemini-2.0-flash":      "deepseek-v4-flash",
		"gemini-2.0-flash-lite": "deepseek-v4-flash",
		"gemini-2.5-pro":        "deepseek-v4-pro",
		"gemini-2.5-flash":      "deepseek-v4-flash",
		"gemini-2.5-flash-lite": "deepseek-v4-flash",
		"gemini-3.1-pro":        "deepseek-v4-pro",
		"gemini-3-pro":          "deepseek-v4-pro",
		"gemini-3-flash":        "deepseek-v4-flash",
		"gemini-3.1-flash":      "deepseek-v4-flash",
		"gemini-3.1-flash-lite": "deepseek-v4-flash",

		"llama-3.1-70b-instruct": "deepseek-v4-flash",
		"qwen-max":               "deepseek-v4-flash",
	}
}

type ModelPolicyError struct {
	Status  int
	Reason  string
	Family  string
	Message string
}

func (e *ModelPolicyError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func ModelPolicyErrorStatus(err error) int {
	if policyErr, ok := err.(*ModelPolicyError); ok && policyErr.Status > 0 {
		return policyErr.Status
	}
	return 400
}

func ResolveModel(store ModelAliasReader, requested string) (string, bool) {
	model, err := ResolveModelOrError(store, requested)
	return model, err == nil
}

func ResolveModelOrError(store ModelAliasReader, requested string) (string, error) {
	model := lower(strings.TrimSpace(requested))
	if model == "" {
		return "", fmt.Errorf("model is empty")
	}
	aliases := loadModelAliases(store)
	if IsSupportedDeepSeekModel(model) {
		return applyModelFamilyPolicy(store, model)
	}
	if mapped, ok := aliases[model]; ok && IsSupportedDeepSeekModel(mapped) {
		return applyModelFamilyPolicy(store, mapped)
	}
	baseModel, noThinking := splitNoThinkingModel(model)
	if mapped, ok := aliases[baseModel]; ok && IsSupportedDeepSeekModel(mapped) {
		return applyModelFamilyPolicy(store, withNoThinkingVariant(mapped, noThinking))
	}
	return "", fmt.Errorf("model %q is not available", requested)
}

func lower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}

func OpenAIModelsResponse() map[string]any {
	return map[string]any{"object": "list", "data": DeepSeekModels}
}

func OpenAIModelByID(store ModelAliasReader, id string) (ModelInfo, bool) {
	canonical, ok := ResolveModel(store, id)
	if !ok {
		return ModelInfo{}, false
	}
	for _, model := range DeepSeekModels {
		if model.ID == canonical {
			return model, true
		}
	}
	return ModelInfo{}, false
}

func OllamaModelsResponse() map[string]any {
	return map[string]any{"models": OllamaModels}
}

func OllamaModelByID(store ModelAliasReader, id string) (OllamaCapabilitiesModelInfo, bool) {
	canonical, ok := ResolveModel(store, id)
	if !ok {
		return OllamaCapabilitiesModelInfo{}, false
	}
	for _, model := range OllamaCapabilitiesModels {
		if model.ID == canonical {
			return model, true
		}
	}
	return OllamaCapabilitiesModelInfo{}, false
}

func ClaudeModelsResponse() map[string]any {
	resp := map[string]any{"object": "list", "data": ClaudeModels}
	if len(ClaudeModels) > 0 {
		resp["first_id"] = ClaudeModels[0].ID
		resp["last_id"] = ClaudeModels[len(ClaudeModels)-1].ID
	} else {
		resp["first_id"] = nil
		resp["last_id"] = nil
	}
	resp["has_more"] = false
	return resp
}

func appendNoThinkingVariants(models []ModelInfo) []ModelInfo {
	out := make([]ModelInfo, 0, len(models)*2)
	for _, model := range models {
		out = append(out, model)
		variant := model
		variant.ID = withNoThinkingVariant(model.ID, true)
		out = append(out, variant)
	}
	return out
}
func mapToOllamaModels(models []ModelInfo) []OllamaModelInfo {
	out := make([]OllamaModelInfo, 0, len(models))
	for _, model := range models {
		var modifiedAt string
		if model.Created > 0 {
			modifiedAt = time.Unix(model.Created, 0).Format(time.RFC3339)
		}
		ollamaModel := OllamaModelInfo{
			Name:       model.ID,
			Model:      model.ID,
			Size:       0,
			ModifiedAt: modifiedAt,
		}
		out = append(out, ollamaModel)
	}
	return out
}

func splitNoThinkingModel(model string) (string, bool) {
	model = lower(strings.TrimSpace(model))
	if strings.HasSuffix(model, noThinkingModelSuffix) {
		return strings.TrimSuffix(model, noThinkingModelSuffix), true
	}
	return model, false
}

func withNoThinkingVariant(model string, enabled bool) string {
	baseModel, _ := splitNoThinkingModel(model)
	if !enabled {
		return baseModel
	}
	if baseModel == "" {
		return ""
	}
	return baseModel + noThinkingModelSuffix
}

func loadModelAliases(store ModelAliasReader) map[string]string {
	aliases := DefaultModelAliases()
	if store != nil {
		for k, v := range store.ModelAliases() {
			aliases[lower(strings.TrimSpace(k))] = lower(strings.TrimSpace(v))
		}
	}
	return aliases
}

func NormalizeModelFamilyPolicy(policy ModelFamilyPolicyConfig) ModelFamilyPolicyConfig {
	normalizeRule := func(rule ModelFamilyPolicyRule) ModelFamilyPolicyRule {
		mode := strings.ToLower(strings.TrimSpace(rule.Mode))
		target := strings.ToLower(strings.TrimSpace(rule.Target))
		if mode == "" {
			mode = "allow"
		}
		if mode != "route" {
			target = ""
		}
		return ModelFamilyPolicyRule{
			Mode:   mode,
			Target: target,
		}
	}
	return ModelFamilyPolicyConfig{
		Flash:  normalizeRule(policy.Flash),
		Pro:    normalizeRule(policy.Pro),
		Vision: normalizeRule(policy.Vision),
	}
}

func applyModelFamilyPolicy(store ModelAliasReader, model string) (string, error) {
	reader, ok := store.(ModelFamilyPolicyReader)
	if !ok || isEmptyModelFamilyPolicy(reader.ModelFamilyPolicy()) {
		return model, nil
	}
	policy := reader.ModelFamilyPolicy()
	baseModel, noThinking := splitNoThinkingModel(model)
	if baseModel == "" {
		return model, nil
	}
	sourceFamily, sourceSearch, ok := modelFamilyForBase(baseModel)
	if !ok {
		return model, nil
	}

	visited := map[string]struct{}{}
	currentFamily := sourceFamily
	for {
		if _, exists := visited[currentFamily]; exists {
			return "", &ModelPolicyError{Message: fmt.Sprintf("model family policy contains route cycle involving %s", currentFamily)}
		}
		visited[currentFamily] = struct{}{}
		rule := modelFamilyRule(policy, currentFamily)
		mode := strings.ToLower(strings.TrimSpace(rule.Mode))
		switch mode {
		case "", "allow":
			return buildModelFromFamily(currentFamily, sourceSearch, noThinking), nil
		case "disable":
			return "", &ModelPolicyError{
				Status:  403,
				Reason:  "disabled",
				Family:  currentFamily,
				Message: fmt.Sprintf("model family %q is disabled by server policy", currentFamily),
			}
		case "route":
			currentFamily = strings.ToLower(strings.TrimSpace(rule.Target))
		default:
			return "", &ModelPolicyError{
				Status:  400,
				Reason:  "invalid_policy",
				Family:  currentFamily,
				Message: fmt.Sprintf("model family %q has invalid mode %q", currentFamily, rule.Mode),
			}
		}
	}
}

func modelFamilyForBase(baseModel string) (family string, search bool, ok bool) {
	switch baseModel {
	case "deepseek-v4-flash":
		return "flash", false, true
	case "deepseek-v4-flash-search":
		return "flash", true, true
	case "deepseek-v4-pro":
		return "pro", false, true
	case "deepseek-v4-pro-search":
		return "pro", true, true
	case "deepseek-v4-vision":
		return "vision", false, true
	default:
		return "", false, false
	}
}

func buildModelFromFamily(family string, search, noThinking bool) string {
	var baseModel string
	switch family {
	case "flash":
		if search {
			baseModel = "deepseek-v4-flash-search"
		} else {
			baseModel = "deepseek-v4-flash"
		}
	case "pro":
		if search {
			baseModel = "deepseek-v4-pro-search"
		} else {
			baseModel = "deepseek-v4-pro"
		}
	case "vision":
		baseModel = "deepseek-v4-vision"
	}
	return withNoThinkingVariant(baseModel, noThinking)
}

func modelFamilyRule(policy ModelFamilyPolicyConfig, family string) ModelFamilyPolicyRule {
	switch family {
	case "flash":
		return policy.Flash
	case "pro":
		return policy.Pro
	case "vision":
		return policy.Vision
	default:
		return ModelFamilyPolicyRule{}
	}
}
