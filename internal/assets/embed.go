package assets

import (
	"embed"
	"io/fs"
)

// Static assets embedded at build time
//go:embed web/templates/* web/static/*
var embeddedFS embed.FS

// Templates contains all HTML templates
//go:embed web/templates/*
var Templates embed.FS

// Static contains all static assets (CSS, JS, images)
//go:embed web/static/*
var Static embed.FS

// GetTemplatesFS returns the embedded templates filesystem
func GetTemplatesFS() fs.FS {
	templatesFS, err := fs.Sub(Templates, "web/templates")
	if err != nil {
		panic("failed to create templates filesystem: " + err.Error())
	}
	return templatesFS
}

// GetStaticFS returns the embedded static assets filesystem
func GetStaticFS() fs.FS {
	staticFS, err := fs.Sub(Static, "web/static")
	if err != nil {
		panic("failed to create static filesystem: " + err.Error())
	}
	return staticFS
}

// GetEmbeddedFS returns the full embedded filesystem
func GetEmbeddedFS() fs.FS {
	return embeddedFS
}