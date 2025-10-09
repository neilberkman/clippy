package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/neilberkman/clippy"
	"github.com/neilberkman/clippy/pkg/recent"
)

// CopyArgs defines arguments for the copy tool
type CopyArgs struct {
	Text      string `json:"text,omitempty" jsonschema:"description=Text content to copy to clipboard"`
	File      string `json:"file,omitempty" jsonschema:"description=File path to copy to clipboard"`
	ForceText string `json:"force_text,omitempty" jsonschema:"description=Set to 'true' to force copying file content as text (only used with 'file' parameter)"`
}

// PasteArgs defines arguments for the paste tool
type PasteArgs struct {
	Destination string `json:"destination,omitempty" jsonschema:"description=Directory to paste files to (defaults to current directory)"`
}

// RecentDownloadsArgs defines arguments for the recent downloads tool
type RecentDownloadsArgs struct {
	Count    int    `json:"count,omitempty" jsonschema:"description=Number of recent files to return (default: 10)"`
	Duration string `json:"duration,omitempty" jsonschema:"description=Time duration to look back (e.g. 5m, 1h)"`
}

// CopyResult defines the result of a copy operation
type CopyResult struct {
	Success bool   `json:"success"`
	Type    string `json:"type,omitempty" jsonschema:"description=Type of content copied (text or file)"`
	Message string `json:"message,omitempty"`
}

// PasteResult defines the result of a paste operation
type PasteResult struct {
	Success bool     `json:"success"`
	Files   []string `json:"files,omitempty" jsonschema:"description=List of files that were pasted"`
	Text    string   `json:"text,omitempty" jsonschema:"description=Text content that was pasted"`
	Message string   `json:"message,omitempty"`
}

// RecentFile represents a recent download
type RecentFile struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
}

// AgentBuffer represents an in-memory clipboard buffer for agent use
// Stores actual file bytes, not generated tokens
type AgentBuffer struct {
	Content     []byte `json:"-"`                 // Raw bytes from file
	Lines       int    `json:"lines,omitempty"`   // Number of lines copied
	SourceFile  string `json:"source_file,omitempty"`
	SourceRange string `json:"source_range,omitempty"` // e.g. "17-23" or "all"
}

// BufferCopyArgs defines arguments for buffer_copy tool
type BufferCopyArgs struct {
	File      string `json:"file" jsonschema:"description=File path to copy from (required)"`
	StartLine int    `json:"start_line,omitempty" jsonschema:"description=Starting line number (1-indexed, omit for entire file)"`
	EndLine   int    `json:"end_line,omitempty" jsonschema:"description=Ending line number (inclusive, omit for entire file)"`
}

// BufferPasteArgs defines arguments for buffer_paste tool
type BufferPasteArgs struct {
	File   string `json:"file" jsonschema:"description=Target file path (required)"`
	Mode   string `json:"mode,omitempty" jsonschema:"description=Paste mode: 'append' (default), 'insert', or 'replace'"`
	AtLine int    `json:"at_line,omitempty" jsonschema:"description=Line number for insert/replace mode (1-indexed)"`
	ToLine int    `json:"to_line,omitempty" jsonschema:"description=End line for replace mode (inclusive, required for replace)"`
}

// BufferResult defines the result of buffer operations
type BufferResult struct {
	Success     bool   `json:"success"`
	Message     string `json:"message,omitempty"`
	Lines       int    `json:"lines,omitempty"`
	SourceFile  string `json:"source_file,omitempty"`
	SourceRange string `json:"source_range,omitempty"`
}

// StartServer starts the MCP server
func StartServer() error {
	// Create MCP server
	s := server.NewMCPServer(
		"Clippy MCP Server",
		"1.0.0",
	)

	// Create agent clipboard buffer (persists for the session)
	// Stores raw file bytes for true copy/paste without token regeneration
	agentBuffer := &AgentBuffer{
		Content: []byte{},
	}

	// Define copy tool
	copyTool := mcp.NewTool(
		"clipboard_copy",
		mcp.WithDescription("Copy text or file to clipboard. CRITICAL: Use 'text' parameter for ANY generated content, code, messages, or text that will be pasted. Use 'file' parameter ONLY for existing files that need to be attached/uploaded. DEFAULT TO 'text' FOR ALL GENERATED CONTENT. PRO TIP: For iterative editing, write to a temp file then use file + force_text='true' to avoid regenerating full content each time."),
		mcp.WithString("text", mcp.Description("Text content to copy - USE THIS for all generated content, code snippets, messages, emails, documentation, or any text that will be pasted")),
		mcp.WithString("file", mcp.Description("File path to copy as file reference - ONLY use this for existing files on disk that need to be dragged/attached, NOT for generated content. PRO TIP: Use with force_text='true' for efficient iterative editing of temp files.")),
		mcp.WithString("force_text", mcp.Description("Set to 'true' to force copying file content as text (only with 'file' parameter). USEFUL PATTERN: Write code to /tmp/script.ext, edit incrementally with Edit tool, then copy with file='/tmp/script.ext' force_text='true' for efficient iterative development without regenerating full text.")),
	)

	// Add copy tool handler
	s.AddTool(copyTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args CopyArgs
		argsBytes, _ := json.Marshal(request.Params.Arguments)
		if err := json.Unmarshal(argsBytes, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}

		// Validate that only one of text or file is provided
		if args.Text != "" && args.File != "" {
			return nil, fmt.Errorf("provide either text or file, not both")
		}

		if args.Text == "" && args.File == "" {
			return nil, fmt.Errorf("provide either text or file to copy")
		}

		var result CopyResult

		if args.Text != "" {
			// Copy text
			err := clippy.CopyText(args.Text)
			if err != nil {
				result = CopyResult{
					Success: false,
					Message: fmt.Sprintf("Failed to copy text: %v", err),
				}
			} else {
				result = CopyResult{
					Success: true,
					Type:    "text",
					Message: fmt.Sprintf("Copied %d characters to clipboard", len(args.Text)),
				}
			}
		} else {
			// Copy file
			absPath, err := filepath.Abs(args.File)
			if err != nil {
				return nil, fmt.Errorf("invalid file path: %w", err)
			}

			// Check if file exists
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				return nil, fmt.Errorf("file not found: %s", absPath)
			}

			forceText := args.ForceText == "true" || args.ForceText == "1"
			copyResult, err := clippy.CopyWithResultAndMode(absPath, forceText)
			if err != nil {
				result = CopyResult{
					Success: false,
					Message: fmt.Sprintf("Failed to copy file: %v", err),
				}
			} else {
				typeStr := "file"
				if copyResult.AsText {
					typeStr = "text content"
				}
				result = CopyResult{
					Success: true,
					Type:    typeStr,
					Message: fmt.Sprintf("Copied %s as %s", filepath.Base(absPath), typeStr),
				}
			}
		}

		// Convert result to JSON for response
		resultJSON, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{
				Type: "text",
				Text: string(resultJSON),
			}},
		}, nil
	})

	// Define paste tool
	pasteTool := mcp.NewTool(
		"clipboard_paste",
		mcp.WithDescription("Paste clipboard content to file or directory. Intelligently handles both text content and file references from clipboard."),
		mcp.WithString("destination", mcp.Description("Destination directory (defaults to current directory)")),
	)

	// Add paste tool handler
	s.AddTool(pasteTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args PasteArgs
		argsBytes, _ := json.Marshal(request.Params.Arguments)
		if err := json.Unmarshal(argsBytes, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}

		// Default to current directory
		if args.Destination == "" {
			args.Destination = "."
		}

		// Try pasting to file
		pasteResult, err := clippy.PasteToFile(args.Destination)
		if err != nil {
			// If that fails, try getting text content
			text, hasText := clippy.GetText()
			if hasText {
				result := PasteResult{
					Success: true,
					Text:    text,
					Message: "Retrieved text from clipboard",
				}
				resultJSON, _ := json.Marshal(result)
				return &mcp.CallToolResult{
					Content: []mcp.Content{mcp.TextContent{
						Type: "text",
						Text: string(resultJSON),
					}},
				}, nil
			}

			result := PasteResult{
				Success: false,
				Message: fmt.Sprintf("Failed to paste: %v", err),
			}
			resultJSON, _ := json.Marshal(result)
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{
					Type: "text",
					Text: string(resultJSON),
				}},
			}, nil
		}

		// Build result based on what was pasted
		result := PasteResult{
			Success: true,
		}

		switch pasteResult.Type {
		case "files":
			result.Files = pasteResult.Files
			result.Message = fmt.Sprintf("Pasted %d file(s) to %s", len(pasteResult.Files), args.Destination)
		case "text":
			if len(pasteResult.Files) > 0 {
				result.Text = pasteResult.Files[0] // Text saved to file
				result.Message = fmt.Sprintf("Pasted text content to %s", pasteResult.Files[0])
			}
		}

		resultJSON, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{
				Type: "text",
				Text: string(resultJSON),
			}},
		}, nil
	})

	// Define recent downloads tool
	recentTool := mcp.NewTool(
		"get_recent_downloads",
		mcp.WithDescription("Get list of recently added files from Downloads, Desktop, and Documents folders. Only shows files that were recently added to these directories."),
		mcp.WithNumber("count", mcp.Description("Number of files to return (default: 10)")),
		mcp.WithString("duration", mcp.Description("Time duration to look back (e.g. 5m, 1h)")),
	)

	// Add recent downloads tool handler
	s.AddTool(recentTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args RecentDownloadsArgs
		argsBytes, _ := json.Marshal(request.Params.Arguments)
		if err := json.Unmarshal(argsBytes, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}

		// Set defaults
		if args.Count == 0 {
			args.Count = 10
		}

		// Parse duration if provided
		config := recent.PickerConfig{}
		if args.Duration != "" {
			maxAge, err := recent.ParseDuration(args.Duration)
			if err != nil {
				return nil, fmt.Errorf("invalid duration: %w", err)
			}
			config.MaxAge = maxAge
		}

		// Get recent downloads
		files, err := recent.GetRecentDownloads(config, args.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to get recent downloads: %w", err)
		}

		// Convert to response format
		var recentFiles []RecentFile
		for _, file := range files {
			recentFiles = append(recentFiles, RecentFile{
				Path:     file.Path,
				Name:     file.Name,
				Size:     file.Size,
				Modified: file.Modified.Format("2006-01-02 15:04:05"),
			})
		}

		resultJSON, _ := json.Marshal(recentFiles)
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{
				Type: "text",
				Text: string(resultJSON),
			}},
		}, nil
	})

	// Define buffer_copy tool
	bufferCopyTool := mcp.NewTool(
		"buffer_copy",
		mcp.WithDescription("Copy file bytes to agent's private buffer. Reads actual file bytes (no token generation). Supports line ranges for precise refactoring. Agent never touches or regenerates the copied content."),
		mcp.WithString("file", mcp.Description("File path to copy from (required)"), mcp.Required()),
		mcp.WithNumber("start_line", mcp.Description("Starting line number (1-indexed, omit for entire file)")),
		mcp.WithNumber("end_line", mcp.Description("Ending line number (inclusive, omit for entire file)")),
	)

	// Add buffer_copy tool handler
	s.AddTool(bufferCopyTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args BufferCopyArgs
		argsBytes, _ := json.Marshal(request.Params.Arguments)
		if err := json.Unmarshal(argsBytes, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}

		if args.File == "" {
			return nil, fmt.Errorf("file parameter is required")
		}

		absPath, err := filepath.Abs(args.File)
		if err != nil {
			return nil, fmt.Errorf("invalid file path: %w", err)
		}

		// Read the entire file
		content, err := os.ReadFile(absPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		lines := strings.Split(string(content), "\n")
		var rangeStr string
		var linesToCopy []string

		// Handle line range
		if args.StartLine > 0 || args.EndLine > 0 {
			start := args.StartLine
			end := args.EndLine

			if start < 1 {
				start = 1
			}
			if end < 1 || end > len(lines) {
				end = len(lines)
			}
			if start > end {
				return nil, fmt.Errorf("start_line (%d) cannot be greater than end_line (%d)", start, end)
			}

			linesToCopy = lines[start-1 : end]
			rangeStr = fmt.Sprintf("%d-%d", start, end)
		} else {
			linesToCopy = lines
			rangeStr = "all"
		}

		// Store raw bytes in buffer
		copiedContent := []byte(strings.Join(linesToCopy, "\n"))
		agentBuffer.Content = copiedContent
		agentBuffer.Lines = len(linesToCopy)
		agentBuffer.SourceFile = filepath.Base(absPath)
		agentBuffer.SourceRange = rangeStr

		result := BufferResult{
			Success:     true,
			Message:     fmt.Sprintf("Copied %d lines from %s (lines %s)", len(linesToCopy), filepath.Base(absPath), rangeStr),
			Lines:       len(linesToCopy),
			SourceFile:  filepath.Base(absPath),
			SourceRange: rangeStr,
		}

		resultJSON, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{
				Type: "text",
				Text: string(resultJSON),
			}},
		}, nil
	})

	// Define buffer_paste tool
	bufferPasteTool := mcp.NewTool(
		"buffer_paste",
		mcp.WithDescription("Paste file bytes from agent's buffer to file. Writes exact bytes without agent token generation. Supports append/insert/replace modes for surgical refactoring."),
		mcp.WithString("file", mcp.Description("Target file path (required)"), mcp.Required()),
		mcp.WithString("mode", mcp.Description("Paste mode: 'append' (default), 'insert', or 'replace'")),
		mcp.WithNumber("at_line", mcp.Description("Line number for insert/replace mode (1-indexed)")),
		mcp.WithNumber("to_line", mcp.Description("End line for replace mode (inclusive, required for replace)")),
	)

	// Add buffer_paste tool handler
	s.AddTool(bufferPasteTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args BufferPasteArgs
		argsBytes, _ := json.Marshal(request.Params.Arguments)
		if err := json.Unmarshal(argsBytes, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}

		if len(agentBuffer.Content) == 0 {
			return nil, fmt.Errorf("buffer is empty - use buffer_copy first")
		}

		if args.File == "" {
			return nil, fmt.Errorf("file parameter is required")
		}

		absPath, err := filepath.Abs(args.File)
		if err != nil {
			return nil, fmt.Errorf("invalid file path: %w", err)
		}

		mode := args.Mode
		if mode == "" {
			mode = "append"
		}

		// Read target file if it exists
		var targetLines []string
		existingContent, err := os.ReadFile(absPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to read target file: %w", err)
			}
			// File doesn't exist, create it
			targetLines = []string{}
		} else {
			targetLines = strings.Split(string(existingContent), "\n")
		}

		bufferLines := strings.Split(string(agentBuffer.Content), "\n")
		var newLines []string

		switch mode {
		case "append":
			// Append buffer content to end of file
			newLines = append(targetLines, bufferLines...)

		case "insert":
			if args.AtLine < 1 {
				return nil, fmt.Errorf("at_line is required for insert mode")
			}
			insertAt := args.AtLine - 1
			if insertAt > len(targetLines) {
				insertAt = len(targetLines)
			}
			// Insert buffer content at specified line
			newLines = make([]string, 0, len(targetLines)+len(bufferLines))
			newLines = append(newLines, targetLines[:insertAt]...)
			newLines = append(newLines, bufferLines...)
			newLines = append(newLines, targetLines[insertAt:]...)

		case "replace":
			if args.AtLine < 1 || args.ToLine < 1 {
				return nil, fmt.Errorf("at_line and to_line are required for replace mode")
			}
			replaceFrom := args.AtLine - 1
			replaceTo := args.ToLine
			if replaceFrom >= len(targetLines) {
				return nil, fmt.Errorf("at_line %d is beyond file length %d", args.AtLine, len(targetLines))
			}
			if replaceTo > len(targetLines) {
				replaceTo = len(targetLines)
			}
			// Replace lines [from, to] with buffer content
			newLines = make([]string, 0)
			newLines = append(newLines, targetLines[:replaceFrom]...)
			newLines = append(newLines, bufferLines...)
			newLines = append(newLines, targetLines[replaceTo:]...)

		default:
			return nil, fmt.Errorf("invalid mode %q: must be 'append', 'insert', or 'replace'", mode)
		}

		// Write the new content
		newContent := []byte(strings.Join(newLines, "\n"))
		if err := os.WriteFile(absPath, newContent, 0644); err != nil {
			return nil, fmt.Errorf("failed to write file: %w", err)
		}

		result := BufferResult{
			Success:     true,
			Message:     fmt.Sprintf("Pasted %d lines to %s (mode: %s)", agentBuffer.Lines, filepath.Base(absPath), mode),
			Lines:       agentBuffer.Lines,
			SourceFile:  agentBuffer.SourceFile,
			SourceRange: agentBuffer.SourceRange,
		}

		resultJSON, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{
				Type: "text",
				Text: string(resultJSON),
			}},
		}, nil
	})

	// Define buffer_list tool
	bufferListTool := mcp.NewTool(
		"buffer_list",
		mcp.WithDescription("Show what's currently in the agent's buffer (metadata only, not content)."),
	)

	// Add buffer_list tool handler
	s.AddTool(bufferListTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if len(agentBuffer.Content) == 0 {
			result := BufferResult{
				Success: true,
				Message: "Buffer is empty",
			}
			resultJSON, _ := json.Marshal(result)
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{
					Type: "text",
					Text: string(resultJSON),
				}},
			}, nil
		}

		result := BufferResult{
			Success:     true,
			Message:     fmt.Sprintf("Buffer contains %d lines from %s (lines %s)", agentBuffer.Lines, agentBuffer.SourceFile, agentBuffer.SourceRange),
			Lines:       agentBuffer.Lines,
			SourceFile:  agentBuffer.SourceFile,
			SourceRange: agentBuffer.SourceRange,
		}

		resultJSON, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{
				Type: "text",
				Text: string(resultJSON),
			}},
		}, nil
	})

	// Add prompts for common operations
	s.AddPrompt(mcp.NewPrompt(
		"copy-recent-download",
		mcp.WithPromptDescription("Copy the most recent download to clipboard"),
		mcp.WithArgument("count"),
	), func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		count := "1"
		if val, ok := request.Params.Arguments["count"]; ok {
			count = fmt.Sprintf("%v", val)
		}

		return &mcp.GetPromptResult{
			Messages: []mcp.PromptMessage{
				{
					Role: mcp.RoleUser,
					Content: mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Copy my %s most recent download(s) to the clipboard.", count),
					},
				},
			},
		}, nil
	})

	s.AddPrompt(mcp.NewPrompt(
		"paste-here",
		mcp.WithPromptDescription("Paste clipboard content to current directory"),
	), func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return &mcp.GetPromptResult{
			Messages: []mcp.PromptMessage{
				{
					Role: mcp.RoleUser,
					Content: mcp.TextContent{
						Type: "text",
						Text: "Paste the clipboard content to the current directory.",
					},
				},
			},
		}, nil
	})

	// Start the server
	return server.ServeStdio(s)
}
