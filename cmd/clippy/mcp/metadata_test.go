package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempJSON(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "override.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp json: %v", err)
	}
	return path
}

func findToolParam(tool ToolSpec, name string) (ToolParamSpec, bool) {
	for _, param := range tool.Params {
		if param.Name == name {
			return param, true
		}
	}
	return ToolParamSpec{}, false
}

func TestDefaultServerMetadataLoads(t *testing.T) {
	metadata, err := DefaultServerMetadata()
	if err != nil {
		t.Fatalf("DefaultServerMetadata: %v", err)
	}
	if len(metadata.Tools) == 0 {
		t.Fatalf("expected default tools")
	}

	toolMap := metadata.ToolMap()
	copyTool, ok := toolMap["clipboard_copy"]
	if !ok {
		t.Fatalf("expected clipboard_copy tool")
	}
	if _, ok := findToolParam(copyTool, "text"); !ok {
		t.Fatalf("expected clipboard_copy text param")
	}
}

func TestLoadServerMetadataPartialToolOverride(t *testing.T) {
	override := `[
  {
    "name": "clipboard_copy",
    "parameters": {
      "properties": {
        "text": {"description": "New text description"}
      }
    }
  }
]`
	path := writeTempJSON(t, override)

	metadata, err := LoadServerMetadata(ServerOptions{ToolsPath: path})
	if err != nil {
		t.Fatalf("LoadServerMetadata: %v", err)
	}

	copyTool := metadata.ToolMap()["clipboard_copy"]
	param, ok := findToolParam(copyTool, "text")
	if !ok {
		t.Fatalf("expected clipboard_copy text param")
	}
	if param.Description != "New text description" {
		t.Fatalf("expected updated description, got %q", param.Description)
	}
}

func TestLoadServerMetadataStrictToolOverrideMissingParams(t *testing.T) {
	override := `[
  {
    "name": "clipboard_copy",
    "description": "Only description"
  }
]`
	path := writeTempJSON(t, override)

	_, err := LoadServerMetadata(ServerOptions{ToolsPath: path, StrictMetadata: true})
	if err == nil {
		t.Fatalf("expected strict metadata error")
	}
}

func TestLoadServerMetadataToolRequiredMismatch(t *testing.T) {
	override := `[
  {
    "name": "buffer_copy",
    "parameters": {
      "required": []
    }
  }
]`
	path := writeTempJSON(t, override)

	_, err := LoadServerMetadata(ServerOptions{ToolsPath: path})
	if err == nil {
		t.Fatalf("expected required mismatch error")
	}
}

func TestLoadServerMetadataPartialPromptOverride(t *testing.T) {
	override := `[
  {
    "name": "copy-recent-download",
    "description": "Copy recent downloads quickly"
  }
]`
	path := writeTempJSON(t, override)

	metadata, err := LoadServerMetadata(ServerOptions{PromptsPath: path})
	if err != nil {
		t.Fatalf("LoadServerMetadata: %v", err)
	}

	prompt := metadata.PromptMap()["copy-recent-download"]
	if prompt.Description != "Copy recent downloads quickly" {
		t.Fatalf("expected updated prompt description, got %q", prompt.Description)
	}
}

func TestLoadServerMetadataStrictPromptOverrideMissingArgs(t *testing.T) {
	override := `[
  {
    "name": "copy-recent-download",
    "description": "Copy recent downloads quickly"
  }
]`
	path := writeTempJSON(t, override)

	_, err := LoadServerMetadata(ServerOptions{PromptsPath: path, StrictMetadata: true})
	if err == nil {
		t.Fatalf("expected strict metadata error")
	}
}
