package settings

import (
	"net/http"
	"strings"

	authn "ds2api/internal/auth"
	"ds2api/internal/config"
	"ds2api/internal/promptcompat"
)

func (h *Handler) getSettings(w http.ResponseWriter, _ *http.Request) {
	snap := h.Store.Snapshot()
	familyPolicy := config.NormalizeModelFamilyPolicy(snap.ModelFamilyPolicy)
	recommended := defaultRuntimeRecommended(len(snap.Accounts), h.Store.RuntimeAccountMaxInflight())
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"admin": map[string]any{
			"has_password_hash":        strings.TrimSpace(snap.Admin.PasswordHash) != "",
			"jwt_expire_hours":         h.Store.AdminJWTExpireHours(),
			"jwt_valid_after_unix":     snap.Admin.JWTValidAfterUnix,
			"default_password_warning": authn.UsingDefaultAdminKey(h.Store),
		},
		"runtime": map[string]any{
			"account_max_inflight":         h.Store.RuntimeAccountMaxInflight(),
			"account_max_queue":            h.Store.RuntimeAccountMaxQueue(recommended),
			"global_max_inflight":          h.Store.RuntimeGlobalMaxInflight(recommended),
			"token_refresh_interval_hours": h.Store.RuntimeTokenRefreshIntervalHours(),
		},
		"responses":   snap.Responses,
		"embeddings":  snap.Embeddings,
		"auto_delete": snap.AutoDelete,
		"current_input_file": map[string]any{
			"flash":  h.Store.CurrentInputFileFlashEnabled(),
			"pro":    h.Store.CurrentInputFileProEnabled(),
			"vision": h.Store.CurrentInputFileVisionEnabled(),
		},
		"model_family_policy": map[string]any{
			"flash":  familyPolicy.Flash,
			"pro":    familyPolicy.Pro,
			"vision": familyPolicy.Vision,
		},
		"thinking_injection": map[string]any{
			"enabled":        h.Store.ThinkingInjectionEnabled(),
			"prompt":         h.Store.ThinkingInjectionPrompt(),
			"default_prompt": promptcompat.DefaultThinkingInjectionPrompt,
		},
		"model_aliases": snap.ModelAliases,
		"env_backed":    h.Store.IsEnvBacked(),
	})
}
