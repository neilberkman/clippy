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
        wantText bool
    }{
        {"text file", "test-files/sample.txt", true},
        {"elixir code", "test-files/code.elixir", true},
        {"pdf file", "test-files/test.pdf", false},
        {"png file", "test-files/minimal.png", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := exec.Command("./clippy_test", "--verbose", tt.file)
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
        {"text file piped", "", true, "test-files/sample.txt"},
        {"binary file piped", "", false, "test-files/test.pdf"},
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
            if tt.wantText {
                if !strings.Contains(outputStr, "Copied text content from stream") {
                    t.Errorf("Expected text stream copy, got: %s", outputStr)
                }
            } else {
                if !strings.Contains(outputStr, "Copied stream as") {
                    t.Errorf("Expected binary stream copy, got: %s", outputStr)
                }
            }
        })
    }
}

func TestFlags(t *testing.T) {
    t.Run("silent by default", func(t *testing.T) {
        cmd := exec.Command("./clippy_test", "test-files/sample.txt")
        output, err := cmd.CombinedOutput()
        if err != nil {
            t.Fatalf("clippy failed: %v", err)
        }
        
        if len(output) > 0 {
            t.Errorf("Expected no output in silent mode, got: %s", output)
        }
    })

    t.Run("verbose flag", func(t *testing.T) {
        cmd := exec.Command("./clippy_test", "--verbose", "test-files/sample.txt")
        output, err := cmd.CombinedOutput()
        if err != nil {
            t.Fatalf("clippy failed: %v", err)
        }
        
        if !strings.Contains(string(output), "✅") {
            t.Errorf("Expected verbose output, got: %s", output)
        }
    })

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
    
    cmd := exec.Command("./clippy_test", "test-files/sample.txt")
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