package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/neilberkman/clippy"
)

func TestCopySlackPreviewAndValidation(t *testing.T) {
	for _, tc := range []struct {
		name     string
		args     map[string]any
		wantErr  bool
		wantCode int
	}{
		{"preview fenced table", map[string]any{"markdown": "```\n  a  b\n  c  d\n```", "preview": true, "expected_code_blocks": 1}, false, 1},
		{"preview without new options", map[string]any{"markdown": "text", "preview": true}, false, 0},
		{"count mismatch", map[string]any{"markdown": "`code`", "expected_code_blocks": 1}, true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var request mcp.CallToolRequest
			request.Params.Arguments = tc.args
			response, err := handleCopySlack(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			if response.IsError != tc.wantErr {
				t.Errorf("IsError = %v, want %v", response.IsError, tc.wantErr)
			}
			var result struct {
				CopyResult
				clippy.SlackCopyResult
			}
			if err := json.Unmarshal([]byte(response.Content[0].(mcp.TextContent).Text), &result); err != nil {
				t.Fatal(err)
			}
			if result.Success == tc.wantErr || result.Copied || result.Formatting.CodeBlocks != tc.wantCode {
				t.Errorf("unexpected result: %+v", result)
			}
			if !strings.Contains(result.Message, "clipboard unchanged") {
				t.Errorf("result did not explain clipboard state: %s", result.Message)
			}
		})
	}
}

func TestCopySlackRejectsInvalidArguments(t *testing.T) {
	for _, args := range []map[string]any{
		{"markdown": " "},
		{"markdown": "text", "preview": "true"},
		{"markdown": "text", "expected_code_blocks": 0.5},
		{"markdown": "text", "expected_code_blocks": -1},
	} {
		var request mcp.CallToolRequest
		request.Params.Arguments = args
		response, err := handleCopySlack(context.Background(), request)
		if err != nil || !response.IsError {
			t.Errorf("args %+v: response = %+v, error = %v", args, response, err)
		}
	}
}
