package deepseek

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"ds2api/internal/auth"
	"ds2api/internal/config"
)

// SessionInfo 会话信息
type SessionInfo struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	TitleType string  `json:"title_type"`
	Pinned    bool    `json:"pinned"`
	UpdatedAt float64 `json:"updated_at"`
}

// SessionStats 会话统计结果
type SessionStats struct {
	AccountID    string  // 账号标识 (email 或 mobile)
	TotalCount   int     // 总会话数量
	PinnedCount  int     // 置顶会话数量
	HasMore      bool    // 是否还有更多
	Success      bool    // 请求是否成功
	ErrorMessage string  // 错误信息
}

// GetSessionCount 获取单个账号的会话数量
func (c *Client) GetSessionCount(ctx context.Context, a *auth.RequestAuth, maxAttempts int) (*SessionStats, error) {
	if maxAttempts <= 0 {
		maxAttempts = c.maxRetries
	}

	stats := &SessionStats{
		AccountID: a.AccountID,
	}

	attempts := 0
	refreshed := false

	for attempts < maxAttempts {
		headers := c.authHeaders(a.DeepSeekToken)

		// 构建请求 URL
		reqURL := DeepSeekFetchSessionURL + "?lte_cursor.pinned=false"

		resp, status, err := c.getJSONWithStatus(ctx, c.regular, reqURL, headers)
		if err != nil {
			config.Logger.Warn("[get_session_count] request error", "error", err, "account", a.AccountID)
			attempts++
			continue
		}

		code := intFrom(resp["code"])
		if status == http.StatusOK && code == 0 {
			data, _ := resp["data"].(map[string]any)
			bizData, _ := data["biz_data"].(map[string]any)
			chatSessions, _ := bizData["chat_sessions"].([]any)
			hasMore, _ := bizData["has_more"].(bool)

			stats.TotalCount = len(chatSessions)
			stats.HasMore = hasMore
			stats.Success = true

			// 统计置顶会话数量
			for _, session := range chatSessions {
				if s, ok := session.(map[string]any); ok {
					if pinned, ok := s["pinned"].(bool); ok && pinned {
						stats.PinnedCount++
					}
				}
			}

			return stats, nil
		}

		msg, _ := resp["msg"].(string)
		stats.ErrorMessage = fmt.Sprintf("status=%d, code=%d, msg=%s", status, code, msg)
		config.Logger.Warn("[get_session_count] failed", "status", status, "code", code, "msg", msg, "account", a.AccountID)

		if a.UseConfigToken {
			if isTokenInvalid(status, code, msg) && !refreshed {
				if c.Auth.RefreshToken(ctx, a) {
					refreshed = true
					continue
				}
			}
			if c.Auth.SwitchAccount(ctx, a) {
				refreshed = false
				attempts++
				continue
			}
		}
		attempts++
	}

	stats.Success = false
	stats.ErrorMessage = "get session count failed after retries"
	return stats, errors.New(stats.ErrorMessage)
}

// GetSessionCountForToken 直接使用 token 获取会话数量（直通模式）
func (c *Client) GetSessionCountForToken(ctx context.Context, token string) (*SessionStats, error) {
	headers := c.authHeaders(token)
	reqURL := DeepSeekFetchSessionURL + "?lte_cursor.pinned=false"

	resp, status, err := c.getJSONWithStatus(ctx, c.regular, reqURL, headers)
	if err != nil {
		return nil, err
	}

	code := intFrom(resp["code"])
	if status != http.StatusOK || code != 0 {
		msg, _ := resp["msg"].(string)
		return nil, fmt.Errorf("request failed: status=%d, code=%d, msg=%s", status, code, msg)
	}

	data, _ := resp["data"].(map[string]any)
	bizData, _ := data["biz_data"].(map[string]any)
	chatSessions, _ := bizData["chat_sessions"].([]any)
	hasMore, _ := bizData["has_more"].(bool)

	stats := &SessionStats{
		TotalCount:  len(chatSessions),
		HasMore:     hasMore,
		Success:     true,
	}

	// 统计置顶会话数量
	for _, session := range chatSessions {
		if s, ok := session.(map[string]any); ok {
			if pinned, ok := s["pinned"].(bool); ok && pinned {
				stats.PinnedCount++
			}
		}
	}

	return stats, nil
}

// GetSessionCountAll 获取所有账号的会话数量统计
func (c *Client) GetSessionCountAll(ctx context.Context) []*SessionStats {
	accounts := c.Store.Accounts()
	results := make([]*SessionStats, 0, len(accounts))

	for _, acc := range accounts {
		token := acc.Token
		accountID := acc.Email
		if accountID == "" {
			accountID = acc.Mobile
		}

		// 如果没有 token，尝试登录获取
		if token == "" {
			var err error
			token, err = c.Login(ctx, acc)
			if err != nil {
				results = append(results, &SessionStats{
					AccountID:    accountID,
					Success:      false,
					ErrorMessage: fmt.Sprintf("login failed: %v", err),
				})
				continue
			}
		}

		stats, err := c.GetSessionCountForToken(ctx, token)
		if err != nil {
			results = append(results, &SessionStats{
				AccountID:    accountID,
				Success:      false,
				ErrorMessage: err.Error(),
			})
			continue
		}

		stats.AccountID = accountID
		results = append(results, stats)
	}

	return results
}

// FetchSessionPage 获取会话列表（支持分页）
func (c *Client) FetchSessionPage(ctx context.Context, a *auth.RequestAuth, cursor string) ([]SessionInfo, bool, error) {
	headers := c.authHeaders(a.DeepSeekToken)

	// 构建请求 URL
	params := url.Values{}
	params.Set("lte_cursor.pinned", "false")
	if cursor != "" {
		params.Set("lte_cursor", cursor)
	}
	reqURL := DeepSeekFetchSessionURL + "?" + params.Encode()

	resp, status, err := c.getJSONWithStatus(ctx, c.regular, reqURL, headers)
	if err != nil {
		return nil, false, err
	}

	code := intFrom(resp["code"])
	if status != http.StatusOK || code != 0 {
		msg, _ := resp["msg"].(string)
		return nil, false, fmt.Errorf("request failed: status=%d, code=%d, msg=%s", status, code, msg)
	}

	data, _ := resp["data"].(map[string]any)
	bizData, _ := data["biz_data"].(map[string]any)
	chatSessions, _ := bizData["chat_sessions"].([]any)
	hasMore, _ := bizData["has_more"].(bool)

	sessions := make([]SessionInfo, 0, len(chatSessions))
	for _, s := range chatSessions {
		if m, ok := s.(map[string]any); ok {
			session := SessionInfo{
				ID:        stringFromMap(m, "id"),
				Title:     stringFromMap(m, "title"),
				TitleType: stringFromMap(m, "title_type"),
				Pinned:    boolFromMap(m, "pinned"),
				UpdatedAt: floatFromMap(m, "updated_at"),
			}
			sessions = append(sessions, session)
		}
	}

	return sessions, hasMore, nil
}

// 辅助函数
func stringFromMap(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func boolFromMap(m map[string]any, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func floatFromMap(m map[string]any, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

// DeleteSessionResult 删除会话结果
type DeleteSessionResult struct {
	SessionID   string // 会话 ID
	Success     bool   // 是否成功
	ErrorMessage string // 错误信息
}

// DeleteSession 删除单个会话
func (c *Client) DeleteSession(ctx context.Context, a *auth.RequestAuth, sessionID string, maxAttempts int) (*DeleteSessionResult, error) {
	if maxAttempts <= 0 {
		maxAttempts = c.maxRetries
	}

	result := &DeleteSessionResult{
		SessionID: sessionID,
	}

	if sessionID == "" {
		result.ErrorMessage = "session_id is required"
		return result, errors.New(result.ErrorMessage)
	}

	attempts := 0
	refreshed := false

	for attempts < maxAttempts {
		headers := c.authHeaders(a.DeepSeekToken)

		payload := map[string]any{
			"chat_session_id": sessionID,
		}

		resp, status, err := c.postJSONWithStatus(ctx, c.regular, DeepSeekDeleteSessionURL, headers, payload)
		if err != nil {
			config.Logger.Warn("[delete_session] request error", "error", err, "session_id", sessionID)
			attempts++
			continue
		}

		code := intFrom(resp["code"])
		if status == http.StatusOK && code == 0 {
			result.Success = true
			return result, nil
		}

		msg, _ := resp["msg"].(string)
		result.ErrorMessage = fmt.Sprintf("status=%d, code=%d, msg=%s", status, code, msg)
		config.Logger.Warn("[delete_session] failed", "status", status, "code", code, "msg", msg, "session_id", sessionID)

		if a.UseConfigToken {
			if isTokenInvalid(status, code, msg) && !refreshed {
				if c.Auth.RefreshToken(ctx, a) {
					refreshed = true
					continue
				}
			}
			if c.Auth.SwitchAccount(ctx, a) {
				refreshed = false
				attempts++
				continue
			}
		}
		attempts++
	}

	result.Success = false
	result.ErrorMessage = "delete session failed after retries"
	return result, errors.New(result.ErrorMessage)
}

// DeleteSessionForToken 直接使用 token 删除会话（直通模式）
func (c *Client) DeleteSessionForToken(ctx context.Context, token string, sessionID string) (*DeleteSessionResult, error) {
	result := &DeleteSessionResult{
		SessionID: sessionID,
	}

	if sessionID == "" {
		result.ErrorMessage = "session_id is required"
		return result, errors.New(result.ErrorMessage)
	}

	headers := c.authHeaders(token)
	payload := map[string]any{
		"chat_session_id": sessionID,
	}

	resp, status, err := c.postJSONWithStatus(ctx, c.regular, DeepSeekDeleteSessionURL, headers, payload)
	if err != nil {
		result.ErrorMessage = err.Error()
		return result, err
	}

	code := intFrom(resp["code"])
	if status != http.StatusOK || code != 0 {
		msg, _ := resp["msg"].(string)
		result.ErrorMessage = fmt.Sprintf("request failed: status=%d, code=%d, msg=%s", status, code, msg)
		return result, fmt.Errorf(result.ErrorMessage)
	}

	result.Success = true
	return result, nil
}

// DeleteAllSessions 删除所有会话（谨慎使用）
func (c *Client) DeleteAllSessions(ctx context.Context, a *auth.RequestAuth) (int, error) {
	deleted := 0
	cursor := ""

	for {
		sessions, hasMore, err := c.FetchSessionPage(ctx, a, cursor)
		if err != nil {
			return deleted, err
		}

		for _, session := range sessions {
			_, err := c.DeleteSession(ctx, a, session.ID, 1)
			if err == nil {
				deleted++
			}
		}

		if !hasMore || len(sessions) == 0 {
			break
		}
	}

	return deleted, nil
}

// DeleteAllSessionsForToken 直接使用 token 删除所有会话（直通模式）
func (c *Client) DeleteAllSessionsForToken(ctx context.Context, token string) (int, error) {
	deleted := 0
	cursor := ""

	for {
		// 获取会话列表
		headers := c.authHeaders(token)
		params := url.Values{}
		params.Set("lte_cursor.pinned", "false")
		if cursor != "" {
			params.Set("lte_cursor", cursor)
		}
		reqURL := DeepSeekFetchSessionURL + "?" + params.Encode()

		resp, status, err := c.getJSONWithStatus(ctx, c.regular, reqURL, headers)
		if err != nil {
			return deleted, err
		}

		code := intFrom(resp["code"])
		if status != http.StatusOK || code != 0 {
			msg, _ := resp["msg"].(string)
			return deleted, fmt.Errorf("fetch sessions failed: status=%d, code=%d, msg=%s", status, code, msg)
		}

		data, _ := resp["data"].(map[string]any)
		bizData, _ := data["biz_data"].(map[string]any)
		chatSessions, _ := bizData["chat_sessions"].([]any)
		hasMore, _ := bizData["has_more"].(bool)

		// 删除每个会话
		for _, s := range chatSessions {
			if m, ok := s.(map[string]any); ok {
				sessionID := stringFromMap(m, "id")
				if sessionID == "" {
					continue
				}
				_, err := c.DeleteSessionForToken(ctx, token, sessionID)
				if err == nil {
					deleted++
				}
			}
		}

		if !hasMore || len(chatSessions) == 0 {
			break
		}
	}

	return deleted, nil
}
