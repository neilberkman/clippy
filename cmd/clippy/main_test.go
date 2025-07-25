package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	// Build the binary for testing
	cmd := exec.Command("go", "build", "-o", "clippy_test", ".")
	if err := cmd.Run(); err != nil {
		panic("Failed to build test binary: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Cleanup
	_ = os.Remove("clippy_test")

	os.Exit(code)
}

func TestFileMode(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		args     []string
		wantText bool
	}{
		{"text file default", "../../test-files/sample.txt", []string{"--verbose"}, false},
		{"text file with -t flag", "../../test-files/sample.txt", []string{"--verbose", "-t"}, true},
		{"elixir file default", "../../test-files/code.exs", []string{"--verbose"}, false},
		{"elixir file with -t flag", "../../test-files/code.exs", []string{"--verbose", "-t"}, true},
		{"pdf file", "../../test-files/test.pdf", []string{"--verbose"}, false},
		{"png file", "../../test-files/minimal.png", []string{"--verbose"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append(tt.args, tt.file)
			cmd := exec.Command("./clippy_test", args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("clippy failed: %v\nOutput: %s", err, output)
			}

			outputStr := string(output)
			if tt.wantText {
				if !strings.Contains(outputStr, "Copied text content") {
					t.Errorf("Expected text content copy, got: %s", outputStr)
				}
			} else {
				if !strings.Contains(outputStr, "Copied file reference") {
					t.Errorf("Expected file reference copy, got: %s", outputStr)
				}
			}
		})
	}
}

func TestStreamMode(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantText  bool
		inputFile string // optional: pipe file content
	}{
		{"text stream", "Hello, World!", true, ""},
		{"text file piped", "", true, "../../test-files/sample.txt"},
		{"binary file piped", "", false, "../../test-files/test.pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./clippy_test", "--verbose")

			var stdin bytes.Buffer
			if tt.inputFile != "" {
				content, err := os.ReadFile(tt.inputFile)
				if err != nil {
					t.Fatalf("Failed to read input file: %v", err)
				}
				stdin.Write(content)
			} else {
				stdin.WriteString(tt.input)
			}

			cmd.Stdin = &stdin
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("clippy failed: %v\nOutput: %s", err, output)
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, "Copied content from stream using smart detection") {
				t.Errorf("Expected smart stream copy, got: %s", outputStr)
			}
		})
	}
}

func TestMultipleFiles(t *testing.T) {
	t.Run("copy multiple files", func(t *testing.T) {
		cmd := exec.Command("./clippy_test", "--verbose", "../../test-files/minimal.png", "../../test-files/sample.txt")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("clippy failed: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "Copied 2 file references") {
			t.Errorf("Expected multiple file copy message, got: %s", outputStr)
		}

		if !strings.Contains(outputStr, "minimal.png") || !strings.Contains(outputStr, "sample.txt") {
			t.Errorf("Expected both file names in output, got: %s", outputStr)
		}
	})
}

func TestFlags(t *testing.T) {
	t.Run("silent by default", func(t *testing.T) {
		// Create a temporary config file that sets verbose=false
		homeDir, _ := os.UserHomeDir()
		configPath := filepath.Join(homeDir, ".clippy.conf")

		// Backup existing config if it exists
		var backupData []byte
		hasBackup := false
		if data, err := os.ReadFile(configPath); err == nil {
			backupData = data
			hasBackup = true
		}

		// Write temporary config
		configContent := `verbose = false`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		defer func() {
			if hasBackup {
				_ = os.WriteFile(configPath, backupData, 0644)
			} else {
				_ = os.Remove(configPath)
			}
		}()

		cmd := exec.Command("./clippy_test", "../../test-files/sample.txt")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("clippy failed: %v", err)
		}

		if len(output) > 0 {
			t.Errorf("Expected no output in silent mode, got: %s", output)
		}
	})

	t.Run("verbose flag", func(t *testing.T) {
		cmd := exec.Command("./clippy_test", "--verbose", "../../test-files/sample.txt")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("clippy failed: %v", err)
		}

		if !strings.Contains(string(output), "✅") {
			t.Errorf("Expected verbose output, got: %s", output)
		}
	})

}

func TestPipelines(t *testing.T) {
	tests := []struct {
		name       string
		pipeline   string
		wantOutput string
		wantError  bool
	}{
		{
			name:       "echo text to clippy",
			pipeline:   `echo "Hello from pipeline" | ./clippy_test -v`,
			wantOutput: "Copied content from stream using smart detection",
		},
		{
			name:       "cat text file to clippy",
			pipeline:   `cat ../../test-files/sample.txt | ./clippy_test -v`,
			wantOutput: "Copied content from stream using smart detection",
		},
		{
			name:       "cat binary file to clippy",
			pipeline:   `cat ../../test-files/test.pdf | ./clippy_test -v`,
			wantOutput: "Copied content from stream using smart detection",
		},
		{
			name:       "find and xargs single file",
			pipeline:   `find ../../test-files -name "sample.txt" | xargs ./clippy_test -v`,
			wantOutput: "Copied file reference",
		},
		{
			name:       "find and xargs multiple files",
			pipeline:   `find ../../test-files \( -name "*.png" -o -name "*.pdf" \) -print0 | xargs -0 ./clippy_test -v`,
			wantOutput: "Copied", // Don't hardcode count - test files may change
		},
		{
			name:       "curl to clippy (simulated)",
			pipeline:   `echo -n "GIF89a" | ./clippy_test -v`,
			wantOutput: "Copied content from stream using smart detection",
		},
		{
			name:       "empty input",
			pipeline:   `echo -n "" | ./clippy_test -v`,
			wantOutput: "Clipboard cleared",
			wantError:  false,
		},
		{
			name:       "multiline text",
			pipeline:   `printf "line1\nline2\nline3" | ./clippy_test -v`,
			wantOutput: "Copied content from stream using smart detection",
		},
		{
			name:       "command output pipe",
			pipeline:   `ls -la ../../test-files | head -3 | ./clippy_test -v`,
			wantOutput: "Copied content from stream using smart detection",
		},
		{
			name:       "base64 encoded binary",
			pipeline:   `base64 -i ../../test-files/minimal.png | ./clippy_test -v`,
			wantOutput: "Copied content from stream using smart detection",
		},
		{
			name:       "json data",
			pipeline:   `echo '{"test": "data"}' | ./clippy_test -v`,
			wantOutput: "Copied content from stream using smart detection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("bash", "-c", tt.pipeline)
			output, err := cmd.CombinedOutput()

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v\nOutput: %s", err, output)
			}

			if tt.wantOutput != "" && !strings.Contains(string(output), tt.wantOutput) {
				t.Errorf("Expected output to contain %q, got: %s", tt.wantOutput, output)
			}
		})
	}
}

func TestConfigFile(t *testing.T) {
	// Create a temporary config file
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".clippy.conf")

	// Backup existing config if it exists
	var backupData []byte
	if data, err := os.ReadFile(configPath); err == nil {
		backupData = data
		defer func() {
			_ = os.WriteFile(configPath, backupData, 0644)
		}()
	}

	// Test verbose config
	configContent := `# Test config
verbose = true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cmd := exec.Command("./clippy_test", "../../test-files/sample.txt")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("clippy failed with config: %v", err)
	}

	if !strings.Contains(string(output), "✅") {
		t.Errorf("Config file verbose=true not working, got: %s", output)
	}

	// Cleanup
	_ = os.Remove(configPath)
}
