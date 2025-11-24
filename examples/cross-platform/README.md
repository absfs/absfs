# Cross-Platform Filesystem Example

This example demonstrates the recommended pattern for using absfs with the OS filesystem across Windows, macOS, and Linux.

## The Pattern

Use Go build tags to create platform-specific filesystem initialization:

- **`filesystem_windows.go`**: Windows-specific setup with drive mapping
- **`filesystem_unix.go`**: Unix/macOS setup (direct passthrough)
- **`main.go`**: Your application code (works everywhere!)

## How It Works

### On Windows

Unix-style paths like `/tmp/absfs-example/test.txt` are automatically mapped to Windows paths like `C:\tmp\absfs-example\test.txt`.

You can specify a different drive:
```bash
go run . --drive=D:
```

### On Unix/macOS

Paths work natively. The drive parameter is ignored for cross-platform compatibility.

## Running the Example

```bash
# Default (C: drive on Windows, / on Unix)
go run .

# Custom drive on Windows (ignored on Unix)
go run . --drive=D:
```

## What It Demonstrates

✅ Same code works on all platforms
✅ Unix-style paths everywhere
✅ Explicit drive selection on Windows when needed
✅ Zero overhead on Unix/macOS
✅ Standard Go build tag pattern

## Applying This to Your Project

Copy the two platform-specific files into your project:

1. Copy `filesystem_windows.go` to your project
2. Copy `filesystem_unix.go` to your project
3. Change the package name to match your project
4. Use `NewFS()` in your application code

That's it! Your code will work identically on all platforms.

## Key Benefits

**Minimal cognitive load:** Copy two files, use `/paths` everywhere

**No platform detection:** Build tags handle it automatically

**Explicit control:** Can override drive on Windows when needed

**Standard pattern:** Uses idiomatic Go build tags
