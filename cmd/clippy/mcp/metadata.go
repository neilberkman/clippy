package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

const (
	paramTypeString = "string"
	paramTypeNumber = "number"
)

// ServerOptions controls optional MCP metadata overrides.
type ServerOptions struct {
	ExamplesPath string
	ToolsPath    string
	PromptsPath  string
}

// ServerMetadata describes the MCP server's tools, prompts, and examples.
type ServerMetadata struct {
	Tools    []ToolSpec
	Prompts  []PromptSpec
	Examples []ExampleSpec
}

// ToolSpec describes a tool and its parameters.
type ToolSpec struct {
	Name        string
	Description string
	Params      []ToolParamSpec
}

// ToolParamSpec describes a tool parameter.
type ToolParamSpec struct {
	Name        string
	Description string
	Type        string
	Required    bool
}

// PromptSpec describes a prompt and its arguments.
type PromptSpec struct {
	Name        string
	Description string
	Arguments   []PromptArgSpec
}

// PromptArgSpec describes a prompt argument.
type PromptArgSpec struct {
	Name        string
	Description string
	Required    bool
}

// ExampleSpec describes a prompt example.
type ExampleSpec struct {
	Prompt      string `json:"prompt"`
	Description string `json:"description"`
}

// DefaultServerMetadata returns the built-in MCP metadata definitions.
func DefaultServerMetadata() ServerMetadata {
	return ServerMetadata{
		Tools: []ToolSpec{
			{
				Name:        "clipboard_copy",
				Description: "Copy text or file to clipboard. CRITICAL: Use 'text' parameter for ANY generated content, code, messages, or text that will be pasted. Use 'file' parameter ONLY for existing files that need to be attached/uploaded. DEFAULT TO 'text' FOR ALL GENERATED CONTENT. PRO TIP: For iterative editing, write to a temp file then use file + force_text='true' to avoid regenerating full content each time.",
				Params: []ToolParamSpec{
					{
						Name:        "text",
						Description: "Text content to copy - USE THIS for all generated content, code snippets, messages, emails, documentation, or any text that will be pasted",
						Type:        paramTypeString,
					},
					{
						Name:        "file",
						Description: "File path to copy as file reference - ONLY use this for existing files on disk that need to be dragged/attached, NOT for generated content. PRO TIP: Use with force_text='true' for efficient iterative editing of temp files.",
						Type:        paramTypeString,
					},
					{
						Name:        "force_text",
						Description: "Set to 'true' to force copying file content as text (only with 'file' parameter). USEFUL PATTERN: Write code to /tmp/script.ext, edit incrementally with Edit tool, then copy with file='/tmp/script.ext' force_text='true' for efficient iterative development without regenerating full text.",
						Type:        paramTypeString,
					},
				},
			},
			{
				Name:        "clipboard_paste",
				Description: "Paste clipboard content to file or directory. Intelligently handles both text content and file references from clipboard.",
				Params: []ToolParamSpec{
					{
						Name:        "destination",
						Description: "Destination directory (defaults to current directory)",
						Type:        paramTypeString,
					},
				},
			},
			{
				Name:        "get_recent_downloads",
				Description: "Get list of recently added files from Downloads, Desktop, and Documents folders. Only shows files that were recently added to these directories.",
				Params: []ToolParamSpec{
					{
						Name:        "count",
						Description: "Number of files to return (default: 10)",
						Type:        paramTypeNumber,
					},
					{
						Name:        "duration",
						Description: "Time duration to look back (e.g. 5m, 1h, 7d, 2 weeks ago, yesterday)",
						Type:        paramTypeString,
					},
				},
			},
			{
				Name:        "buffer_copy",
				Description: "Copy file bytes (with optional line ranges) to agent's private buffer for refactoring. Server reads bytes directly - no token generation. Use when moving code between files or reorganizing large sections. Better than Edit for large blocks since content isn't regenerated.",
				Params: []ToolParamSpec{
					{
						Name:        "file",
						Description: "File path to copy from (required)",
						Type:        paramTypeString,
						Required:    true,
					},
					{
						Name:        "start_line",
						Description: "Starting line number (1-indexed, omit for entire file)",
						Type:        paramTypeNumber,
					},
					{
						Name:        "end_line",
						Description: "Ending line number (inclusive, omit for entire file)",
						Type:        paramTypeNumber,
					},
				},
			},
			{
				Name:        "buffer_paste",
				Description: "Paste buffered bytes to file with append/insert/replace modes. Use after buffer_copy to complete refactoring. Writes exact bytes without token generation. append=add to end, insert=inject at line, replace=overwrite range. Byte-perfect, no content regeneration.",
				Params: []ToolParamSpec{
					{
						Name:        "file",
						Description: "Target file path (required)",
						Type:        paramTypeString,
						Required:    true,
					},
					{
						Name:        "mode",
						Description: "Paste mode: 'append' (default), 'insert', or 'replace'",
						Type:        paramTypeString,
					},
					{
						Name:        "at_line",
						Description: "Line number for insert/replace mode (1-indexed)",
						Type:        paramTypeNumber,
					},
					{
						Name:        "to_line",
						Description: "End line for replace mode (inclusive, required for replace)",
						Type:        paramTypeNumber,
					},
				},
			},
			{
				Name:        "buffer_cut",
				Description: "Cut lines from file to buffer - copies to buffer then deletes from source. Like buffer_copy but removes the lines after copying. Use for moving code sections without manual deletion. Atomic operation - only deletes if copy succeeds.",
				Params: []ToolParamSpec{
					{
						Name:        "file",
						Description: "File path to cut from (required)",
						Type:        paramTypeString,
						Required:    true,
					},
					{
						Name:        "start_line",
						Description: "Starting line number (1-indexed, omit for entire file)",
						Type:        paramTypeNumber,
					},
					{
						Name:        "end_line",
						Description: "Ending line number (inclusive, omit for entire file)",
						Type:        paramTypeNumber,
					},
				},
			},
			{
				Name:        "buffer_list",
				Description: "Show buffer metadata (lines, source file, range). Use to verify buffer contents before pasting. Returns metadata only, not actual content.",
			},
		},
		Prompts: []PromptSpec{
			{
				Name:        "copy-recent-download",
				Description: "Copy the most recent download to clipboard",
				Arguments: []PromptArgSpec{
					{
						Name:        "count",
						Description: "Number of recent downloads to copy",
					},
				},
			},
			{
				Name:        "paste-here",
				Description: "Paste clipboard content to current directory",
			},
		},
		Examples: []ExampleSpec{
			{
				Prompt:      "Write a Python script to process CSV files and copy it to my clipboard",
				Description: "Generate code and put it directly on your clipboard",
			},
			{
				Prompt:      "Draft an email about the meeting and put it on my clipboard",
				Description: "Create formatted text ready to paste into any email client",
			},
			{
				Prompt:      "Copy my most recent download to the clipboard",
				Description: "Quickly grab recently downloaded files",
			},
			{
				Prompt:      "Refactor the processData function into a separate file",
				Description: "Use buffer_copy and buffer_paste to move code without touching system clipboard",
			},
		},
	}
}

// LoadServerMetadata loads default metadata and applies any overrides.
func LoadServerMetadata(opts ServerOptions) (ServerMetadata, error) {
	metadata := DefaultServerMetadata()

	if opts.ToolsPath != "" {
		overrides, err := loadToolsOverride(opts.ToolsPath)
		if err != nil {
			return ServerMetadata{}, err
		}
		tools, err := applyToolOverrides(metadata.Tools, overrides)
		if err != nil {
			return ServerMetadata{}, err
		}
		metadata.Tools = tools
	}

	if opts.PromptsPath != "" {
		overrides, err := loadPromptsOverride(opts.PromptsPath)
		if err != nil {
			return ServerMetadata{}, err
		}
		prompts, err := applyPromptOverrides(metadata.Prompts, overrides)
		if err != nil {
			return ServerMetadata{}, err
		}
		metadata.Prompts = prompts
	}

	if opts.ExamplesPath != "" {
		overrides, err := loadExamplesOverride(opts.ExamplesPath)
		if err != nil {
			return ServerMetadata{}, err
		}
		metadata.Examples = overrides
	}

	return metadata, nil
}

func (m ServerMetadata) ToolMap() map[string]ToolSpec {
	result := make(map[string]ToolSpec, len(m.Tools))
	for _, tool := range m.Tools {
		result[tool.Name] = tool
	}
	return result
}

func (m ServerMetadata) PromptMap() map[string]PromptSpec {
	result := make(map[string]PromptSpec, len(m.Prompts))
	for _, prompt := range m.Prompts {
		result[prompt.Name] = prompt
	}
	return result
}

func toolParamDescription(tool ToolSpec, name string) (string, error) {
	for _, param := range tool.Params {
		if param.Name == name {
			return param.Description, nil
		}
	}
	return "", fmt.Errorf("tool %q missing parameter metadata for %q", tool.Name, name)
}

func promptArgSpec(prompt PromptSpec, name string) (PromptArgSpec, error) {
	for _, arg := range prompt.Arguments {
		if arg.Name == name {
			return arg, nil
		}
	}
	return PromptArgSpec{}, fmt.Errorf("prompt %q missing argument metadata for %q", prompt.Name, name)
}

func requireToolSpec(toolSpecs map[string]ToolSpec, name string) (ToolSpec, error) {
	if spec, ok := toolSpecs[name]; ok {
		return spec, nil
	}
	return ToolSpec{}, fmt.Errorf("missing tool metadata for %q", name)
}

func requirePromptSpec(promptSpecs map[string]PromptSpec, name string) (PromptSpec, error) {
	if spec, ok := promptSpecs[name]; ok {
		return spec, nil
	}
	return PromptSpec{}, fmt.Errorf("missing prompt metadata for %q", name)
}

type toolOverride struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Parameters  toolOverrideParams   `json:"parameters"`
}

type toolOverrideParams struct {
	Properties map[string]toolOverrideProperty `json:"properties"`
	Required   []string                        `json:"required"`
}

type toolOverrideProperty struct {
	Description string `json:"description"`
}

type promptOverride struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Arguments   []promptArgOverride  `json:"arguments"`
}

type promptArgOverride struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    *bool  `json:"required"`
}

func loadToolsOverride(path string) ([]toolOverride, error) {
	data, err := readJSONFile(path)
	if err != nil {
		return nil, err
	}

	var list []toolOverride
	if err := json.Unmarshal(data, &list); err == nil && len(list) > 0 {
		return list, nil
	}

	var wrapper struct {
		Tools []toolOverride `json:"tools"`
	}
	if err := json.Unmarshal(data, &wrapper); err == nil && len(wrapper.Tools) > 0 {
		return wrapper.Tools, nil
	}

	return nil, fmt.Errorf("tools override file %s must be a JSON array of tools or an object with a non-empty \"tools\" field", path)
}

func loadPromptsOverride(path string) ([]promptOverride, error) {
	data, err := readJSONFile(path)
	if err != nil {
		return nil, err
	}

	var list []promptOverride
	if err := json.Unmarshal(data, &list); err == nil && len(list) > 0 {
		return list, nil
	}

	var wrapper struct {
		Prompts []promptOverride `json:"prompts"`
	}
	if err := json.Unmarshal(data, &wrapper); err == nil && len(wrapper.Prompts) > 0 {
		return wrapper.Prompts, nil
	}

	return nil, fmt.Errorf("prompts override file %s must be a JSON array of prompts or an object with a non-empty \"prompts\" field", path)
}

func readJSONFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("%s is empty", path)
	}
	return data, nil
}

func applyToolOverrides(defaults []ToolSpec, overrides []toolOverride) ([]ToolSpec, error) {
	if len(overrides) == 0 {
		return nil, fmt.Errorf("tools override file contains no tools")
	}

	defaultMap := make(map[string]ToolSpec, len(defaults))
	for _, tool := range defaults {
		defaultMap[tool.Name] = tool
	}

	overrideMap := make(map[string]toolOverride, len(overrides))
	for _, tool := range overrides {
		if tool.Name == "" {
			return nil, fmt.Errorf("tools override contains a tool with no name")
		}
		if _, exists := overrideMap[tool.Name]; exists {
			return nil, fmt.Errorf("tools override contains duplicate tool %q", tool.Name)
		}
		overrideMap[tool.Name] = tool
	}

	for name := range overrideMap {
		if _, ok := defaultMap[name]; !ok {
			return nil, fmt.Errorf("tools override contains unknown tool %q", name)
		}
	}

	updated := make([]ToolSpec, 0, len(defaults))
	for _, tool := range defaults {
		override, ok := overrideMap[tool.Name]
		if !ok {
			return nil, fmt.Errorf("tools override missing tool %q", tool.Name)
		}
		if override.Description == "" {
			return nil, fmt.Errorf("tools override tool %q missing description", tool.Name)
		}

		paramMap := override.Parameters.Properties
		if paramMap == nil {
			paramMap = map[string]toolOverrideProperty{}
		}

		defaultParamNames := make(map[string]ToolParamSpec, len(tool.Params))
		for _, param := range tool.Params {
			defaultParamNames[param.Name] = param
			if _, ok := paramMap[param.Name]; !ok {
				return nil, fmt.Errorf("tools override tool %q missing parameter %q", tool.Name, param.Name)
			}
			if paramMap[param.Name].Description == "" {
				return nil, fmt.Errorf("tools override tool %q parameter %q missing description", tool.Name, param.Name)
			}
		}

		for name := range paramMap {
			if _, ok := defaultParamNames[name]; !ok {
				return nil, fmt.Errorf("tools override tool %q contains unknown parameter %q", tool.Name, name)
			}
		}

		updatedTool := tool
		updatedTool.Description = override.Description
		for idx := range updatedTool.Params {
			param := updatedTool.Params[idx]
			if overrideParam, ok := paramMap[param.Name]; ok {
				param.Description = overrideParam.Description
				updatedTool.Params[idx] = param
			}
		}
		updated = append(updated, updatedTool)
	}

	return updated, nil
}

func applyPromptOverrides(defaults []PromptSpec, overrides []promptOverride) ([]PromptSpec, error) {
	if len(overrides) == 0 {
		return nil, fmt.Errorf("prompts override file contains no prompts")
	}

	defaultMap := make(map[string]PromptSpec, len(defaults))
	for _, prompt := range defaults {
		defaultMap[prompt.Name] = prompt
	}

	overrideMap := make(map[string]promptOverride, len(overrides))
	for _, prompt := range overrides {
		if prompt.Name == "" {
			return nil, fmt.Errorf("prompts override contains a prompt with no name")
		}
		if _, exists := overrideMap[prompt.Name]; exists {
			return nil, fmt.Errorf("prompts override contains duplicate prompt %q", prompt.Name)
		}
		overrideMap[prompt.Name] = prompt
	}

	for name := range overrideMap {
		if _, ok := defaultMap[name]; !ok {
			return nil, fmt.Errorf("prompts override contains unknown prompt %q", name)
		}
	}

	updated := make([]PromptSpec, 0, len(defaults))
	for _, prompt := range defaults {
		override, ok := overrideMap[prompt.Name]
		if !ok {
			return nil, fmt.Errorf("prompts override missing prompt %q", prompt.Name)
		}
		if override.Description == "" {
			return nil, fmt.Errorf("prompts override prompt %q missing description", prompt.Name)
		}

		argMap := make(map[string]promptArgOverride, len(override.Arguments))
		for _, arg := range override.Arguments {
			if arg.Name == "" {
				return nil, fmt.Errorf("prompts override prompt %q contains an argument with no name", prompt.Name)
			}
			if _, exists := argMap[arg.Name]; exists {
				return nil, fmt.Errorf("prompts override prompt %q contains duplicate argument %q", prompt.Name, arg.Name)
			}
			argMap[arg.Name] = arg
		}

		defaultArgNames := make(map[string]PromptArgSpec, len(prompt.Arguments))
		for _, arg := range prompt.Arguments {
			defaultArgNames[arg.Name] = arg
			overrideArg, ok := argMap[arg.Name]
			if !ok {
				return nil, fmt.Errorf("prompts override prompt %q missing argument %q", prompt.Name, arg.Name)
			}
			if overrideArg.Description == "" {
				return nil, fmt.Errorf("prompts override prompt %q argument %q missing description", prompt.Name, arg.Name)
			}
			if overrideArg.Required != nil && *overrideArg.Required != arg.Required {
				return nil, fmt.Errorf("prompts override prompt %q argument %q required mismatch", prompt.Name, arg.Name)
			}
		}

		for name := range argMap {
			if _, ok := defaultArgNames[name]; !ok {
				return nil, fmt.Errorf("prompts override prompt %q contains unknown argument %q", prompt.Name, name)
			}
		}

		updatedPrompt := prompt
		updatedPrompt.Description = override.Description
		for idx := range updatedPrompt.Arguments {
			arg := updatedPrompt.Arguments[idx]
			if overrideArg, ok := argMap[arg.Name]; ok {
				arg.Description = overrideArg.Description
				updatedPrompt.Arguments[idx] = arg
			}
		}
		updated = append(updated, updatedPrompt)
	}

	return updated, nil
}

func validateExamples(examples []ExampleSpec) error {
	for idx, example := range examples {
		if example.Prompt == "" {
			return fmt.Errorf("examples override entry %d missing prompt", idx+1)
		}
		if example.Description == "" {
			return fmt.Errorf("examples override entry %d missing description", idx+1)
		}
	}
	return nil
}

func loadExamplesOverride(path string) ([]ExampleSpec, error) {
	examples, err := loadExamplesOverrideFile(path)
	if err != nil {
		return nil, err
	}
	if err := validateExamples(examples); err != nil {
		return nil, err
	}
	return examples, nil
}

func loadExamplesOverrideFile(path string) ([]ExampleSpec, error) {
	data, err := readJSONFile(path)
	if err != nil {
		return nil, err
	}

	var list []ExampleSpec
	if err := json.Unmarshal(data, &list); err == nil && len(list) > 0 {
		return list, nil
	}

	var wrapper struct {
		Examples []ExampleSpec `json:"examples"`
	}
	if err := json.Unmarshal(data, &wrapper); err == nil && len(wrapper.Examples) > 0 {
		return wrapper.Examples, nil
	}

	return nil, fmt.Errorf("examples override file %s must be a JSON array of examples or an object with a non-empty \"examples\" field", path)
}
