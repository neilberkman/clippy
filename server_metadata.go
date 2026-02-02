package clippy

import _ "embed"

// DefaultServerJSON contains the MCP server metadata defaults.
//
//go:embed server.json
var DefaultServerJSON []byte
