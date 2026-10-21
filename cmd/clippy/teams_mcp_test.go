package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/neilberkman/clippy/cmd/clippy/mcp"
	"github.com/neilberkman/clippy/pkg/clipboard"
)

func TestCopyTeamsMCP(t *testing.T) {
	for _, overrides := range []bool{false, true} {
		name := "defaults"
		if overrides {
			name = "overrides"
		}
		t.Run(name, func(t *testing.T) {
			metadata, err := mcpserver.DefaultServerMetadata()
			if err != nil {
				t.Fatal(err)
			}
			spec := metadata.ToolMap()["copy_teams"]
			args := []string{"mcp-server"}
			if overrides {
				path := filepath.Join(t.TempDir(), "tools.json")
				if err := os.WriteFile(path, []byte(`[{"name":"copy_teams","description":"Teams override","parameters":{"properties":{"markdown":{"description":"Markdown override"}}}}]`), 0o600); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--tools", path)
				spec.Description = "Teams override"
				spec.Params[0].Description = "Markdown override"
			}
			client, err := mcpclient.NewStdioMCPClient("./clippy_test", nil, args...)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = client.Close() }()
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			var initialize mcp.InitializeRequest
			initialize.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
			initialize.Params.ClientInfo = mcp.Implementation{Name: "clippy-test", Version: "1"}
			if _, err := client.Initialize(ctx, initialize); err != nil {
				t.Fatal(err)
			}
			if err := client.Ping(ctx); err != nil {
				t.Fatal(err)
			}
			tools, err := client.ListTools(ctx, mcp.ListToolsRequest{})
			if err != nil {
				t.Fatal(err)
			}
			index := slices.IndexFunc(tools.Tools, func(tool mcp.Tool) bool { return tool.Name == "copy_teams" })
			if index < 0 {
				t.Fatal("copy_teams is not registered")
			}
			tool := tools.Tools[index]
			property, ok := tool.InputSchema.Properties["markdown"].(map[string]any)
			if !ok || property["type"] != "string" || property["description"] != spec.Params[0].Description || tool.Description != spec.Description || len(tool.InputSchema.Properties) != 1 || !slices.Equal(tool.InputSchema.Required, []string{"markdown"}) {
				t.Fatalf("runtime schema does not match metadata: %+v", tool)
			}
			annotations := tool.Annotations
			if annotations.Title != "Copy Teams-ready message" || annotations.ReadOnlyHint == nil || *annotations.ReadOnlyHint || annotations.DestructiveHint == nil || *annotations.DestructiveHint || annotations.IdempotentHint == nil || !*annotations.IdempotentHint || annotations.OpenWorldHint == nil || *annotations.OpenWorldHint {
				t.Errorf("unexpected annotations: %+v", annotations)
			}
			for _, arguments := range []map[string]any{nil, {"markdown": " \t\n"}, {"markdown": 123}} {
				if err := clipboard.CopyText("keep this"); err != nil {
					t.Fatal(err)
				}
				var request mcp.CallToolRequest
				request.Params.Name = "copy_teams"
				request.Params.Arguments = arguments
				response, err := client.CallTool(ctx, request)
				if err == nil && !response.IsError {
					t.Errorf("invalid arguments succeeded: %+v", arguments)
				}
				if text, ok := clipboard.GetText(); !ok || text != "keep this" {
					t.Errorf("invalid arguments changed clipboard: %q, %v", text, ok)
				}
			}
			var request mcp.CallToolRequest
			request.Params.Name = "copy_teams"
			request.Params.Arguments = map[string]any{"markdown": "**世界 👋**"}
			response, err := client.CallTool(ctx, request)
			if err != nil {
				t.Fatal(err)
			}
			if response.IsError || len(response.Content) != 1 {
				t.Fatalf("copy_teams failed: %+v", response)
			}
			var result mcpserver.CopyResult
			if err := json.Unmarshal([]byte(response.Content[0].(mcp.TextContent).Text), &result); err != nil {
				t.Fatal(err)
			}
			if !result.Success || result.Type != "rich_teams" || result.Message != "Copied 8 characters as Teams-ready rich text; paste with Command-V (Teams rendering not verified)" {
				t.Errorf("unexpected result: %+v", result)
			}
			html, ok := clipboard.GetClipboardDataForType("public.html")
			if !ok || !strings.HasPrefix(string(html), "<!DOCTYPE html>") || !strings.Contains(string(html), "<b>世界 👋</b>") {
				t.Errorf("tool did not write Teams HTML: %s", html)
			}
			if _, ok := clipboard.GetClipboardDataForType("public.rtf"); !ok {
				t.Error("tool did not write RTF fallback")
			}
		})
	}
}
