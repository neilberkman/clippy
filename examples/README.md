# Clippy Library Examples

This directory contains examples of how to use clippy as a Go library in your own applications.

## Running the Examples

```bash
# From the examples directory
cd examples
go run example_simple.go
```

## Example Files

### `example_simple.go`

A basic example showing the main library functions:

- `clippy.CopyText()` - Copy text to clipboard
- `clippy.Copy()` - Smart copy (detects text vs binary files)
- `clippy.CopyData()` - Copy from reader with automatic text/binary detection
- `clippy.GetText()` - Get text from clipboard
- `clippy.GetFiles()` - Get file paths from clipboard

This example demonstrates both the high-level API (with smart detection) and basic clipboard operations.

## Using in Your Projects

Add clippy to your project:

```bash
go get github.com/neilberkman/clippy
```

Then import and use:

```go
import "github.com/neilberkman/clippy"

// Your code here
err := clippy.Copy("document.pdf")
```

See the main README for more detailed API documentation.
