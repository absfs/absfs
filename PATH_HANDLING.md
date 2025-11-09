# Path Handling in absfs

## Overview

The `absfs` package provides a **virtual filesystem abstraction** that works consistently across platforms (Unix, macOS, Windows). Understanding how paths are interpreted is crucial for writing portable code.

## Path Semantics

### Virtual Absolute Paths

In `absfs`, paths starting with `/` or `\` are treated as **virtual absolute paths**, even on Windows where they aren't technically OS-absolute (which require a drive letter like `C:\`).

```go
fs := absfs.ExtendFiler(myFiler)

// These work identically on all platforms:
fs.Create("/config/app.json")      // ✅ Virtual absolute
fs.Open("/data/users.db")          // ✅ Virtual absolute
fs.MkdirAll("/var/log", 0755)      // ✅ Virtual absolute
```

### OS Absolute Paths

True OS-absolute paths are also supported and work as expected:

```go
// Unix/macOS - These are OS-absolute AND virtual-absolute
fs.Open("/usr/local/bin/app")      // ✅ Same as above

// Windows - These are OS-absolute (filepath.IsAbs returns true)
fs.Open("C:\\Program Files\\app")  // ✅ Windows drive letter path
fs.Open("\\\\server\\share\\file") // ✅ Windows UNC path
```

## Platform-Specific Behavior

| Path Type | Unix/macOS | Windows | Notes |
|-----------|------------|---------|-------|
| `/path/to/file` | OS-absolute ✅ | Virtual-absolute ⚠️ | Works everywhere |
| `\path\to\file` | Relative | Virtual-absolute ⚠️ | Unix treats `\` as filename char |
| `C:\path` | Relative | OS-absolute ✅ | Windows-specific |
| `\\server\share` | Relative | OS-absolute UNC ✅ | Windows UNC |
| `//server/share` | OS-absolute | Virtual-absolute ⚠️ | Unix UNC-style |
| `relative/path` | Relative | Relative | Platform-independent |

⚠️ = Virtual-absolute (works in absfs but not truly OS-absolute on Windows)

## Design Rationale

This design allows **virtual filesystems** (mocks, in-memory, archives) to use Unix-style paths universally while maintaining compatibility with OS filesystems:

### Virtual Filesystems (Primary Use Case)
```go
// Mock filesystem for testing
type MockFS struct { /* ... */ }

func TestMyCode(t *testing.T) {
    fs := absfs.ExtendFiler(&MockFS{})

    // Same test code works on Windows, macOS, Linux
    fs.Create("/test/data.txt")      // ✅ Portable
    fs.MkdirAll("/tmp/cache", 0755)  // ✅ Portable
}
```

### OS Filesystems (Still Supported)
```go
// Wrapping actual OS operations
import "github.com/absfs/osfs"

fs := osfs.NewFS()

// OS-native paths work correctly
fs.Open("C:\\Windows\\file.txt")    // ✅ Windows
fs.Open("/usr/local/bin/app")       // ✅ Unix/macOS
```

## Best Practices

### ✅ DO: Use Unix-style paths for virtual filesystems
```go
// Portable across all platforms
fs := absfs.ExtendFiler(virtualFS)
fs.Create("/config/settings.json")
fs.MkdirAll("/var/log/app", 0755)
```

### ✅ DO: Use OS-native paths when wrapping the real filesystem
```go
import "github.com/absfs/osfs"

fs := osfs.NewFS()

// Windows
fs.Open("C:\\Users\\Documents\\file.txt")

// Unix/macOS
fs.Open("/home/user/documents/file.txt")
```

### ⚠️ CAUTION: Mixing virtual and OS paths
```go
// On Windows, this might not work as expected
fs.Open("/Windows/System32/config")  // Virtual-absolute, not C:\Windows\...

// Better: Use OS-native paths for real filesystem access
fs.Open("C:\\Windows\\System32\\config")  // OS-absolute
```

### ❌ DON'T: Assume `/path` is OS-absolute on Windows
```go
// This creates a file at \myapp\data.txt (relative to current drive)
// NOT C:\myapp\data.txt
osFS.Create("/myapp/data.txt")  // ⚠️ Surprising on Windows

// Better: Use full path with drive letter
osFS.Create("C:\\myapp\\data.txt")  // ✅ Explicit
```

## UNC Path Support

Windows UNC paths (network shares) are fully supported:

```go
// Proper Windows UNC (OS-absolute)
fs.Open("\\\\server\\share\\file.txt")  // ✅ Windows

// Unix-style UNC (virtual-absolute on Windows)
fs.Open("//server/share/file.txt")      // ✅ Cross-platform virtual FS
```

## Path Conversion

If you need to convert between path styles:

```go
import "path/filepath"

// Platform-independent path joining
path := filepath.Join("/", "config", "app.json")
// Unix/macOS: /config/app.json
// Windows:    \config\app.json

// Clean paths for consistency
cleaned := filepath.Clean("/path/../to/file")
// Result: /to/file (or \to\file on Windows)
```

## Testing Across Platforms

Write tests that work everywhere by using virtual filesystems:

```go
func TestFileOperations(t *testing.T) {
    // Use memfs or mock filesystem
    fs := absfs.ExtendFiler(newMockFiler())

    // Use Unix-style paths - portable to Windows
    err := fs.MkdirAll("/test/data", 0755)
    if err != nil {
        t.Fatal(err)
    }

    f, err := fs.Create("/test/data/file.txt")
    if err != nil {
        t.Fatal(err)
    }
    defer f.Close()

    // Test passes on all platforms ✅
}
```

## Implementation Details

The `isVirtualAbs()` helper determines if a path should be treated as absolute:

```go
func isVirtualAbs(path string) bool {
    // Check OS-absolute first (handles C:\, \\server\share on Windows)
    if filepath.IsAbs(path) {
        return true
    }
    // Treat paths starting with separator as virtual-absolute
    if len(path) > 0 && (path[0] == '/' || path[0] == '\\') {
        return true
    }
    return false
}
```

This ensures:
- ✅ Unix absolute paths work everywhere
- ✅ Windows drive letters work on Windows
- ✅ Windows UNC paths work on Windows
- ✅ Virtual filesystems use consistent paths across platforms

## Summary

- **Virtual filesystems**: Use Unix-style paths (`/path/to/file`) for portability
- **OS filesystems**: Use platform-native paths for correctness
- **Testing**: Virtual filesystems with Unix paths = portable tests
- **Windows**: True absolute paths need drive letters (`C:\`) or UNC (`\\server\share`)
- **Cross-platform**: The abstraction handles the differences automatically

For most users building virtual/mock filesystems for testing, simply use Unix-style paths and everything will work across platforms.
