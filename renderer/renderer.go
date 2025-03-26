package renderer

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin/render"

	"github.com/a-h/templ"
)

// HTMLTemplRenderer is a custom renderer that can handle templ.Component.
type HTMLTemplRenderer struct {
	FallbackHtmlRenderer render.HTMLRender
}

// Instance implements gin's HTMLRender interface. If the data is a templ.Component,
// it returns our custom Renderer struct; otherwise, it falls back to the default
// provided gin HTML renderer (if set).
func (r *HTMLTemplRenderer) Instance(s string, d interface{}) render.Render {
	templData, ok := d.(templ.Component)
	if !ok {
		if r.FallbackHtmlRenderer != nil {
			return r.FallbackHtmlRenderer.Instance(s, d)
		}
	}
	return &Renderer{
		Ctx:       context.Background(),
		Status:    -1,
		Component: templData,
	}
}

// New creates a new Renderer with a given context, status code, and templ component.
func New(ctx context.Context, status int, component templ.Component) *Renderer {
	return &Renderer{
		Ctx:       ctx,
		Status:    status,
		Component: component,
	}
}

// Renderer implements gin.Render for templ components.
type Renderer struct {
	Ctx       context.Context
	Status    int
	Component templ.Component
}

// Render writes headers and renders the templ component to the HTTP response.
func (t Renderer) Render(w http.ResponseWriter) error {
	t.WriteContentType(w)
	if t.Status != -1 {
		w.WriteHeader(t.Status)
	}
	if t.Component != nil {
		return t.Component.Render(t.Ctx, w)
	}
	return nil
}

// WriteContentType sets "Content-Type: text/html; charset=utf-8".
func (t Renderer) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}
