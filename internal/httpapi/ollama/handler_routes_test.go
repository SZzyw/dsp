package ollama

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

type ollamaTestSurface struct {
	Store       shared.ConfigReader
	handler     *Handler
}

GetVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"version":"0.23.1"}`))	
}
func (h *Handler) ListOllamaModels(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, config.OllamaModelsResponse())
}
func (h *Handler) GetOllamaModel

func registerOllamaTestRoutes(r chi.Router, h *ollamaTestSurface) {
	r.Get("/api/version", h.handler().GetVersion)
	r.Get("/api/tags", h.modelsHandler().ListOllamaModels)
	r.Post("/api/show", h.chatHandler().GetOllamaModel)
}


func TestGetOllamaVersionRoute(t *testing.T) {
	h := &ollamaTestSurface{}
	r := chi.NewRouter()
	registerOllamaTestRoutes(r, h)
    req := httptest.NewRequest(http.MethodGet, "/api/version", nil)    
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}


func TestGetOllamaModelsRoute(t *testing.T) {
	h := &ollamaTestSurface{}
	r := chi.NewRouter()
	registerOllamaTestRoutes(r, h)
    req := httptest.NewRequest(http.MethodGet, "/api/tags", nil)    
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}


func TestGetOllamaModelRoute(t *testing.T) {
	h := &ollamaTestSurface{}
	r := chi.NewRouter()
	registerOllamaTestRoutes(r, h)

	t.Run("direct", func(t *testing.T) {
		body := `{"model":"deepseek-v4-flash"}`
    	req := httptest.NewRequest(http.MethodPost, "/api/show", strings.NewReader(body))    
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("direct_nothinking", func(t *testing.T) {
		body := `{"model":"deepseek-v4-flash-nothinking"}`
    	req := httptest.NewRequest(http.MethodPost, "/api/show", strings.NewReader(body))    
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("direct_expert", func(t *testing.T) {
		body := `{"model":"models/deepseek-v4-pro"}`
    	req := httptest.NewRequest(http.MethodPost, "/api/show", strings.NewReader(body))    
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("direct_vision", func(t *testing.T) {
		body := `{"model":"deepseek-v4-vision"}`
    	req := httptest.NewRequest(http.MethodPost, "/api/show", strings.NewReader(body))    
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
		}
	})
}

func TestGetOllamaModelRouteNotFound(t *testing.T) {
	h := &ollamaTestSurface{}
	r := chi.NewRouter()
	registerOllamaTestRoutes(r, h)

	body := `{"model":"not-exists"}`
    req := httptest.NewRequest(http.MethodPost, "/api/show", strings.NewReader(body))    
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}
