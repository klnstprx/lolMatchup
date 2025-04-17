// Package static embeds static assets for the application.
package static

import "embed"

// FS contains the embedded static assets.
//go:embed htmx/*
var FS embed.FS