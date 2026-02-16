package webui

import (
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"

	root "ds2api"
	"ds2api/internal/config"
)

const welcomeHTML = `<!DOCTYPE html>
<html lang="zh-CN"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>DS2API</title>
<style>body{font-family:Inter,system-ui,sans-serif;background:#030712;color:#f9fafb;display:flex;min-height:100vh;align-items:center;justify-content:center;margin:0}a{color:#f59e0b;text-decoration:none}main{max-width:700px;padding:24px;text-align:center}h1{font-size:48px;margin:0 0 12px}.links{display:flex;gap:16px;justify-content:center;margin-top:20px;flex-wrap:wrap}</style>
</head><body><main><h1>DS2API</h1><p>DeepSeek to OpenAI & Claude Compatible API</p><div class="links"><a href="/admin">管理面板</a><a href="/v1/models">API 状态</a><a href="https://github.com/CJackHwang/ds2api" target="_blank">GitHub</a></div></main></body></html>`

type Handler struct {
	StaticDir     string
	embeddedAdmin fs.FS
	hasEmbeddedUI bool
}

func NewHandler() *Handler {
	h := &Handler{StaticDir: config.StaticAdminDir()}
	if sub, err := fs.Sub(root.EmbeddedAdminFS, "static/admin"); err == nil {
		h.embeddedAdmin = sub
		h.hasEmbeddedUI = true
	}
	return h
}

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Get("/", h.index)
	r.Get("/admin", h.admin)
}

func (h *Handler) HandleAdminFallback(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodGet {
		return false
	}
	if !strings.HasPrefix(r.URL.Path, "/admin/") {
		return false
	}
	h.admin(w, r)
	return true
}

func (h *Handler) index(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(welcomeHTML))
}

func (h *Handler) admin(w http.ResponseWriter, r *http.Request) {
	if fi, err := os.Stat(h.StaticDir); err == nil && fi.IsDir() {
		h.serveFromDisk(w, r)
		return
	}
	if h.hasEmbeddedUI {
		if h.serveFromFS(w, r, h.embeddedAdmin) {
			return
		}
	}
	http.Error(w, "WebUI not built. Run `cd webui && npm run build` first.", http.StatusNotFound)
}

func (h *Handler) serveFromDisk(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin")
	path = strings.TrimPrefix(path, "/")
	if path != "" && strings.Contains(path, ".") {
		full := filepath.Join(h.StaticDir, filepath.Clean(path))
		if !strings.HasPrefix(full, h.StaticDir) {
			http.NotFound(w, r)
			return
		}
		if _, err := os.Stat(full); err == nil {
			if strings.HasPrefix(path, "assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				w.Header().Set("Cache-Control", "no-store, must-revalidate")
			}
			http.ServeFile(w, r, full)
			return
		}
		http.NotFound(w, r)
		return
	}
	index := filepath.Join(h.StaticDir, "index.html")
	if _, err := os.Stat(index); err != nil {
		http.Error(w, "index.html not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Cache-Control", "no-store, must-revalidate")
	http.ServeFile(w, r, index)
}

func (h *Handler) serveFromFS(w http.ResponseWriter, r *http.Request, rootFS fs.FS) bool {
	rel := strings.TrimPrefix(r.URL.Path, "/admin")
	rel = strings.TrimPrefix(rel, "/")
	safe := strings.TrimPrefix(path.Clean("/"+rel), "/")
	if strings.HasPrefix(safe, "../") {
		http.NotFound(w, r)
		return true
	}

	if safe != "" && strings.Contains(safe, ".") {
		if _, err := fs.Stat(rootFS, safe); err != nil {
			http.NotFound(w, r)
			return true
		}
		if strings.HasPrefix(safe, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-store, must-revalidate")
		}
		http.ServeFileFS(w, r, rootFS, safe)
		return true
	}

	if _, err := fs.Stat(rootFS, "index.html"); err != nil {
		return false
	}
	w.Header().Set("Cache-Control", "no-store, must-revalidate")
	http.ServeFileFS(w, r, rootFS, "index.html")
	return true
}
