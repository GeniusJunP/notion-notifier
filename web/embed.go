package web

import "embed"

// DistFS contains the built SPA files from web/dist/.
// Build the SPA with: cd web && npm run build
//
//go:embed all:dist
var DistFS embed.FS
