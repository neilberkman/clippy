package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/neilberkman/clippy"
)

func handleCopySlack(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args CopySlackArgs
	argsBytes, err := json.Marshal(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid arguments: %v", err)), nil
	}
	if err := json.Unmarshal(argsBytes, &args); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid arguments: %v", err)), nil
	}
	if strings.TrimSpace(args.Markdown) == "" {
		return mcp.NewToolResultError("markdown is required"), nil
	}

	report, copyErr := clippy.CopySlackWithOptions(args.Markdown, clippy.SlackCopyOptions{
		Preview:            args.Preview,
		ExpectedCodeBlocks: args.ExpectedCodeBlocks,
	})
	result := struct {
		CopyResult
		*clippy.SlackCopyResult
	}{
		CopyResult:      CopyResult{Success: copyErr == nil, Type: "slack_message"},
		SlackCopyResult: report,
	}
	switch {
	case copyErr != nil:
		result.Message = copyErr.Error()
	case args.Preview:
		result.Message = "Preview only; clipboard unchanged. Counts describe generated formatting, not a verified Slack paste."
	default:
		result.Message = fmt.Sprintf("Copied %d characters as a native Slack message; paste with Command-V. Counts describe generated formatting, not a verified Slack paste. Plain-text clipboard reads omit formatting and fence markers.", utf8.RuneCountInString(args.Markdown))
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("encode Slack copy result: %w", err)
	}
	toolResult := mcp.NewToolResultText(string(resultJSON))
	toolResult.IsError = copyErr != nil
	return toolResult, nil
}
