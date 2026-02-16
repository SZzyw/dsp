package ds2api

import "embed"

// EmbeddedAdminFS bundles static/admin so serverless targets can serve WebUI
// without relying on runtime filesystem inclusion.
//
//go:embed static/admin
var EmbeddedAdminFS embed.FS
