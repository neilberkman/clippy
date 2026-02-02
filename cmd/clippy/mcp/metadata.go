package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/neilberkman/clippy"
)

// ServerOptions controls optional MCP metadata overrides.
type ServerOptions struct {
	ExamplesPath   string
	ToolsPath      string
	PromptsPath    string
	StrictMetadata bool
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

type serverJSON struct {
	Tools    []serverTool  `json:"tools"`
	Prompts  []PromptSpec  `json:"prompts"`
	Examples []ExampleSpec `json:"examples"`
}

type serverTool struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Parameters  serverToolParams `json:"parameters"`
}

type serverToolParams struct {
	Properties map[string]serverToolParam `json:"properties"`
	Required   []string                   `json:"required"`
}

type serverToolParam struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// DefaultServerMetadata returns the built-in MCP metadata definitions.
func DefaultServerMetadata() (ServerMetadata, error) {
	return loadServerMetadataFromJSON(clippy.DefaultServerJSON)
}

func loadServerMetadataFromJSON(data []byte) (ServerMetadata, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return ServerMetadata{}, fmt.Errorf("default server metadata is empty")
	}

	var payload serverJSON
	if err := json.Unmarshal(data, &payload); err != nil {
		return ServerMetadata{}, fmt.Errorf("parse default server metadata: %w", err)
	}
	if len(payload.Tools) == 0 {
		return ServerMetadata{}, fmt.Errorf("default server metadata is missing tools")
	}

	tools := make([]ToolSpec, 0, len(payload.Tools))
	for _, tool := range payload.Tools {
		if tool.Name == "" {
			return ServerMetadata{}, fmt.Errorf("default tool is missing name")
		}
		if strings.TrimSpace(tool.Description) == "" {
			return ServerMetadata{}, fmt.Errorf("default tool %q is missing description", tool.Name)
		}

		requiredSet := make(map[string]bool, len(tool.Parameters.Required))
		for _, name := range tool.Parameters.Required {
			requiredSet[name] = true
		}

		params := make([]ToolParamSpec, 0, len(tool.Parameters.Properties))
		for name, param := range tool.Parameters.Properties {
			if name == "" {
				return ServerMetadata{}, fmt.Errorf("default tool %q has a parameter with no name", tool.Name)
			}
			if strings.TrimSpace(param.Description) == "" {
				return ServerMetadata{}, fmt.Errorf("default tool %q parameter %q missing description", tool.Name, name)
			}
			if strings.TrimSpace(param.Type) == "" {
				return ServerMetadata{}, fmt.Errorf("default tool %q parameter %q missing type", tool.Name, name)
			}
			params = append(params, ToolParamSpec{
				Name:        name,
				Description: param.Description,
				Type:        param.Type,
				Required:    requiredSet[name],
			})
		}
		tools = append(tools, ToolSpec{
			Name:        tool.Name,
			Description: tool.Description,
			Params:      params,
		})
	}

	return ServerMetadata{
		Tools:    tools,
		Prompts:  payload.Prompts,
		Examples: payload.Examples,
	}, nil
}

// LoadServerMetadata loads default metadata and applies any overrides.
func LoadServerMetadata(opts ServerOptions) (ServerMetadata, error) {
	metadata, err := DefaultServerMetadata()
	if err != nil {
		return ServerMetadata{}, err
	}

	if opts.ToolsPath != "" {
		overrides, err := loadToolsOverride(opts.ToolsPath)
		if err != nil {
			return ServerMetadata{}, err
		}
		tools, err := applyToolOverrides(metadata.Tools, overrides, opts.StrictMetadata)
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
		prompts, err := applyPromptOverrides(metadata.Prompts, overrides, opts.StrictMetadata)
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
	Name        string              `json:"name"`
	Description *string             `json:"description,omitempty"`
	Parameters  *toolOverrideParams `json:"parameters,omitempty"`
}

type toolOverrideParams struct {
	Properties map[string]toolOverrideProperty `json:"properties"`
	Required   *[]string                       `json:"required"`
}

type toolOverrideProperty struct {
	Description *string `json:"description,omitempty"`
}

type promptOverride struct {
	Name        string               `json:"name"`
	Description *string              `json:"description,omitempty"`
	Arguments   *[]promptArgOverride `json:"arguments,omitempty"`
}

type promptArgOverride struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Required    *bool   `json:"required,omitempty"`
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

func applyToolOverrides(defaults []ToolSpec, overrides []toolOverride, strict bool) ([]ToolSpec, error) {
	if len(overrides) == 0 {
		return nil, fmt.Errorf("tools override file contains no tools")
	}

	updated := make([]ToolSpec, len(defaults))
	copy(updated, defaults)

	defaultIndex := make(map[string]int, len(defaults))
	defaultMap := make(map[string]ToolSpec, len(defaults))
	for idx, tool := range defaults {
		defaultIndex[tool.Name] = idx
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

	if strict {
		for _, tool := range defaults {
			if _, ok := overrideMap[tool.Name]; !ok {
				return nil, fmt.Errorf("tools override missing tool %q", tool.Name)
			}
		}
	}

	for _, override := range overrides {
		idx := defaultIndex[override.Name]
		tool := updated[idx]

		if override.Description != nil {
			if strings.TrimSpace(*override.Description) == "" {
				return nil, fmt.Errorf("tools override tool %q missing description", tool.Name)
			}
			tool.Description = *override.Description
		} else if strict {
			return nil, fmt.Errorf("tools override tool %q missing description", tool.Name)
		}

		if override.Parameters != nil {
			if override.Parameters.Required != nil {
				if !stringSetEqual(toolRequiredNames(tool), *override.Parameters.Required) {
					return nil, fmt.Errorf("tools override tool %q required parameters mismatch", tool.Name)
				}
			}
			if override.Parameters.Properties != nil {
				paramIndex := toolParamIndex(tool.Params)
				for name, overrideParam := range override.Parameters.Properties {
					paramIdx, ok := paramIndex[name]
					if !ok {
						return nil, fmt.Errorf("tools override tool %q contains unknown parameter %q", tool.Name, name)
					}
					if overrideParam.Description == nil || strings.TrimSpace(*overrideParam.Description) == "" {
						return nil, fmt.Errorf("tools override tool %q parameter %q missing description", tool.Name, name)
					}
					param := tool.Params[paramIdx]
					param.Description = *overrideParam.Description
					tool.Params[paramIdx] = param
				}
			}
		} else if strict && len(tool.Params) > 0 {
			return nil, fmt.Errorf("tools override tool %q missing parameters", tool.Name)
		}

		if strict {
			if override.Parameters == nil || override.Parameters.Properties == nil {
				return nil, fmt.Errorf("tools override tool %q missing parameters", tool.Name)
			}
			for _, param := range tool.Params {
				if _, ok := override.Parameters.Properties[param.Name]; !ok {
					return nil, fmt.Errorf("tools override tool %q missing parameter %q", tool.Name, param.Name)
				}
			}
		}

		updated[idx] = tool
	}

	return updated, nil
}

func applyPromptOverrides(defaults []PromptSpec, overrides []promptOverride, strict bool) ([]PromptSpec, error) {
	if len(overrides) == 0 {
		return nil, fmt.Errorf("prompts override file contains no prompts")
	}

	updated := make([]PromptSpec, len(defaults))
	copy(updated, defaults)

	defaultIndex := make(map[string]int, len(defaults))
	defaultMap := make(map[string]PromptSpec, len(defaults))
	for idx, prompt := range defaults {
		defaultIndex[prompt.Name] = idx
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

	if strict {
		for _, prompt := range defaults {
			if _, ok := overrideMap[prompt.Name]; !ok {
				return nil, fmt.Errorf("prompts override missing prompt %q", prompt.Name)
			}
		}
	}

	for _, override := range overrides {
		idx := defaultIndex[override.Name]
		prompt := updated[idx]

		if override.Description != nil {
			if strings.TrimSpace(*override.Description) == "" {
				return nil, fmt.Errorf("prompts override prompt %q missing description", prompt.Name)
			}
			prompt.Description = *override.Description
		} else if strict {
			return nil, fmt.Errorf("prompts override prompt %q missing description", prompt.Name)
		}

		if override.Arguments != nil {
			argMap := make(map[string]promptArgOverride, len(*override.Arguments))
			for _, arg := range *override.Arguments {
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
			}

			for name, overrideArg := range argMap {
				defaultArg, ok := defaultArgNames[name]
				if !ok {
					return nil, fmt.Errorf("prompts override prompt %q contains unknown argument %q", prompt.Name, name)
				}
				if overrideArg.Description == nil || strings.TrimSpace(*overrideArg.Description) == "" {
					return nil, fmt.Errorf("prompts override prompt %q argument %q missing description", prompt.Name, name)
				}
				if overrideArg.Required != nil && *overrideArg.Required != defaultArg.Required {
					return nil, fmt.Errorf("prompts override prompt %q argument %q required mismatch", prompt.Name, name)
				}
			}

			for idx := range prompt.Arguments {
				arg := prompt.Arguments[idx]
				if overrideArg, ok := argMap[arg.Name]; ok {
					arg.Description = *overrideArg.Description
					prompt.Arguments[idx] = arg
				}
			}
		} else if strict && len(prompt.Arguments) > 0 {
			return nil, fmt.Errorf("prompts override prompt %q missing arguments", prompt.Name)
		}

		if strict {
			if override.Arguments == nil && len(prompt.Arguments) > 0 {
				return nil, fmt.Errorf("prompts override prompt %q missing arguments", prompt.Name)
			}
			if override.Arguments != nil {
				requiredArgs := make(map[string]bool, len(prompt.Arguments))
				for _, arg := range prompt.Arguments {
					requiredArgs[arg.Name] = true
				}
				for _, arg := range *override.Arguments {
					delete(requiredArgs, arg.Name)
				}
				for name := range requiredArgs {
					return nil, fmt.Errorf("prompts override prompt %q missing argument %q", prompt.Name, name)
				}
			}
		}

		updated[idx] = prompt
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

func toolRequiredNames(tool ToolSpec) []string {
	names := make([]string, 0)
	for _, param := range tool.Params {
		if param.Required {
			names = append(names, param.Name)
		}
	}
	return names
}

func toolParamIndex(params []ToolParamSpec) map[string]int {
	index := make(map[string]int, len(params))
	for idx, param := range params {
		index[param.Name] = idx
	}
	return index
}

func stringSetEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	counts := make(map[string]int, len(a))
	for _, name := range a {
		counts[name]++
	}
	for _, name := range b {
		if counts[name] == 0 {
			return false
		}
		counts[name]--
	}
	for _, count := range counts {
		if count != 0 {
			return false
		}
	}
	return true
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
