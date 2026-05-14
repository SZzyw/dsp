package settings

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	authn "ds2api/internal/auth"
	"ds2api/internal/config"
)

func (h *Handler) updateSettings(w http.ResponseWriter, r *http.Request) {
	var req map[string]any
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"detail": "invalid json"})
		return
	}

	adminCfg, runtimeCfg, responsesCfg, embeddingsCfg, autoDeleteCfg, currentInputCfg, modelFamilyCfg, modelToolCfg, aliasMap, err := parseSettingsUpdateRequest(req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"detail": err.Error()})
		return
	}
	if runtimeCfg != nil {
		if err := validateMergedRuntimeSettings(h.Store.Snapshot().Runtime, runtimeCfg); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"detail": err.Error()})
			return
		}
	}
	currentInputFlashSet := hasNestedSettingsKey(req, "current_input_file", "flash")
	currentInputProSet := hasNestedSettingsKey(req, "current_input_file", "pro")
	currentInputVisionSet := hasNestedSettingsKey(req, "current_input_file", "vision")
	modelFamilyFlashSet := hasNestedSettingsKey(req, "model_family_policy", "flash")
	modelFamilyProSet := hasNestedSettingsKey(req, "model_family_policy", "pro")
	modelFamilyVisionSet := hasNestedSettingsKey(req, "model_family_policy", "vision")
	modelToolFlashSet := hasNestedSettingsKey(req, "model_tool_policy", "flash")
	modelToolProSet := hasNestedSettingsKey(req, "model_tool_policy", "pro")
	modelToolVisionSet := hasNestedSettingsKey(req, "model_tool_policy", "vision")

	if err := h.Store.Update(func(c *config.Config) error {
		if adminCfg != nil {
			if adminCfg.JWTExpireHours > 0 {
				c.Admin.JWTExpireHours = adminCfg.JWTExpireHours
			}
		}
		if runtimeCfg != nil {
			if runtimeCfg.AccountMaxInflight > 0 {
				c.Runtime.AccountMaxInflight = runtimeCfg.AccountMaxInflight
			}
			if runtimeCfg.AccountMaxQueue > 0 {
				c.Runtime.AccountMaxQueue = runtimeCfg.AccountMaxQueue
			}
			if runtimeCfg.GlobalMaxInflight > 0 {
				c.Runtime.GlobalMaxInflight = runtimeCfg.GlobalMaxInflight
			}
			if runtimeCfg.TokenRefreshIntervalHours > 0 {
				c.Runtime.TokenRefreshIntervalHours = runtimeCfg.TokenRefreshIntervalHours
			}
			if strings.TrimSpace(runtimeCfg.AccountScheduleMode) != "" {
				c.Runtime.AccountScheduleMode = runtimeCfg.AccountScheduleMode
			}
			if runtimeCfg.AccountStickyReuseCount > 0 {
				c.Runtime.AccountStickyReuseCount = runtimeCfg.AccountStickyReuseCount
			}
		}
		if responsesCfg != nil && responsesCfg.StoreTTLSeconds > 0 {
			c.Responses.StoreTTLSeconds = responsesCfg.StoreTTLSeconds
		}
		if embeddingsCfg != nil && strings.TrimSpace(embeddingsCfg.Provider) != "" {
			c.Embeddings.Provider = strings.TrimSpace(embeddingsCfg.Provider)
		}
		if autoDeleteCfg != nil {
			c.AutoDelete.Mode = autoDeleteCfg.Mode
			c.AutoDelete.Sessions = autoDeleteCfg.Sessions
		}
		if currentInputCfg != nil {
			if currentInputFlashSet {
				c.CurrentInputFile.Flash = currentInputCfg.Flash
			}
			if currentInputProSet {
				c.CurrentInputFile.Pro = currentInputCfg.Pro
			}
			if currentInputVisionSet {
				c.CurrentInputFile.Vision = currentInputCfg.Vision
			}
		}
		if modelFamilyCfg != nil {
			if modelFamilyFlashSet {
				c.ModelFamilyPolicy.Flash = modelFamilyCfg.Flash
			}
			if modelFamilyProSet {
				c.ModelFamilyPolicy.Pro = modelFamilyCfg.Pro
			}
			if modelFamilyVisionSet {
				c.ModelFamilyPolicy.Vision = modelFamilyCfg.Vision
			}
		}
		if modelToolCfg != nil {
			if modelToolFlashSet {
				c.ModelToolPolicy.Flash = modelToolCfg.Flash
			}
			if modelToolProSet {
				c.ModelToolPolicy.Pro = modelToolCfg.Pro
			}
			if modelToolVisionSet {
				c.ModelToolPolicy.Vision = modelToolCfg.Vision
			}
		}
		if aliasMap != nil {
			c.ModelAliases = aliasMap
		}
		return nil
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"detail": err.Error()})
		return
	}

	h.applyRuntimeSettings()
	writeJSON(w, http.StatusOK, map[string]any{
		"success":    true,
		"message":    "settings updated and hot reloaded",
		"env_backed": h.Store.IsEnvBacked(),
	})
}

func (h *Handler) updateSettingsPassword(w http.ResponseWriter, r *http.Request) {
	var req map[string]any
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"detail": "invalid json"})
		return
	}
	newPassword := strings.TrimSpace(fieldString(req, "new_password"))
	if newPassword == "" {
		newPassword = strings.TrimSpace(fieldString(req, "password"))
	}
	if len(newPassword) < 4 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"detail": "new password must be at least 4 characters"})
		return
	}

	now := time.Now().Unix()
	hash := authn.HashAdminPassword(newPassword)
	if err := h.Store.Update(func(c *config.Config) error {
		c.Admin.PasswordHash = hash
		c.Admin.JWTValidAfterUnix = now
		return nil
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"detail": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":              true,
		"message":              "password updated",
		"force_relogin":        true,
		"jwt_valid_after_unix": now,
	})
}

func hasNestedSettingsKey(req map[string]any, section, key string) bool {
	raw, ok := req[section].(map[string]any)
	if !ok {
		return false
	}
	_, exists := raw[key]
	return exists
}
