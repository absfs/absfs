# Path Handling in absfs

**Audience**: All users of absfs - covers cross-platform path semantics and best practices.

**Quick Start:** For OS filesystems on Windows, use the [cross-platform setup pattern](#cross-platform-os-filesystem-setup). Virtual filesystems work with `/paths` on all platforms automatically.

---

## Virtual vs OS Filesystems

### Virtual Filesystems

Virtual filesystems (memfs, s3fs, custom implementations) work with Unix-style paths on all platforms automatically:

```go
import "github.com/absfs/memfs"

fs := memfs.NewFS()

// Works identically on Windows, macOS, and Linux
fs.Create("/config/app.json")
fs.MkdirAll("/var/log/app", 0755)
fs.Open("/data/users.db")
```

**No special setup needed** - just use `/paths` everywhere.

### OS Filesystems

OS filesystems (osfs) interact with your actual operating system. On Windows, you need explicit drive mapping for Unix-style paths to work correctly.

**Use the cross-platform setup pattern below.**

---

## Cross-Platform OS Filesystem Setup

For applications using the OS filesystem across platforms, use Go build tags to handle Windows drive mapping automatically:

### Step 1: Create `filesystem_windows.go`

```go
//go:build windows

package yourapp

import "github.com/absfs/osfs"

// NewFS creates an OS filesystem with drive mapping on Windows
func NewFS(drive string) absfs.FileSystem {
	if drive == "" {
		drive = "C:"  // Default to C: drive
	}
	return osfs.NewWindowsDriveMapper(osfs.NewFS(), drive)
}
```

### Step 2: Create `filesystem_unix.go`

```go
//go:build !windows

package yourapp

import "github.com/absfs/osfs"

// NewFS creates an OS filesystem (no drive mapping needed on Unix)
func NewFS(drive string) absfs.FileSystem {
	return osfs.NewFS()  // Drive parameter ignored on Unix
}
```

### Step 3: Use It Everywhere

```go
package main

import "yourapp"

func main() {
	// Default drive (C: on Windows, / on Unix)
	fs := yourapp.NewFS("")

	// Now write code once, works everywhere!
	fs.Create("/config/app.json")       // C:\config\app.json on Windows
	fs.MkdirAll("/var/log/app", 0755)   // C:\var\log\app on Windows
	fs.Open("/data/users.db")           // C:\data\users.db on Windows
}

// Custom drive on Windows
func withCustomDrive() {
	fs := yourapp.NewFS("D:")  // Maps to D: drive on Windows
	fs.Create("/data/file.txt")  // Creates D:\data\file.txt
}
```

**That's it!** The build tags select the right implementation automatically, and your application code works identically on all platforms.

---

## How It Works

### On Windows

The `WindowsDriveMapper` translates Unix-style paths to Windows paths:

```
/config/app.json  →  C:\config\app.json
/var/log/app      →  C:\var\log\app
/data/users.db    →  C:\data\users.db
```

OS-absolute paths (with drive letters or UNC) pass through unchanged:

```
C:\Windows\file.txt      →  C:\Windows\file.txt  (no change)
D:\Data\file.txt         →  D:\Data\file.txt     (no change)
\\server\share\file.txt  →  \\server\share\file.txt  (no change)
```

### On Unix/macOS

The filesystem is created directly without any wrapper - Unix-style paths work natively:

```
/config/app.json  →  /config/app.json  (native)
/var/log/app      →  /var/log/app      (native)
/data/users.db    →  /data/users.db    (native)
```

---

## When to Use This Pattern

### ✅ Use cross-platform setup for:

- CLI tools that work with OS filesystem
- Applications that need Unix-style paths on Windows
- Code that should work identically on all platforms
- When porting Unix-based tools to Windows

### ✅ Don't need this pattern for:

- Virtual filesystems (memfs, s3fs, etc.) - they already work everywhere
- Testing with mock filesystems - use `/paths` directly
- Applications that only run on one platform
- When you want full control over Windows drive letters in your code

---

## Path Semantics Reference

For those who want deeper understanding, here's how absfs interprets paths:

### Virtual-Absolute Paths

Paths starting with `/` or `\` are treated as **virtual-absolute** - they work in virtual filesystems on all platforms:

```go
fs := memfs.NewFS()

// These work identically everywhere:
fs.Create("/config/app.json")      // ✅ Virtual absolute
fs.Open("/data/users.db")          // ✅ Virtual absolute
fs.MkdirAll("/var/log", 0755)      // ✅ Virtual absolute
```

### OS-Absolute Paths

Platform-specific absolute paths:

```go
// Unix/macOS - These are OS-absolute
fs.Open("/usr/local/bin/app")      // ✅ Native Unix path

// Windows - These are OS-absolute
fs.Open("C:\\Program Files\\app")  // ✅ Windows drive letter
fs.Open("\\\\server\\share\\file") // ✅ Windows UNC path
```

### Platform Behavior Table

| Path Type | Unix/macOS | Windows (without mapper) | Windows (with mapper) |
|-----------|------------|--------------------------|----------------------|
| `/path/to/file` | OS-absolute ✅ | Virtual-absolute ⚠️ | Mapped to `C:\path\to\file` ✅ |
| `C:\path` | Relative | OS-absolute ✅ | OS-absolute ✅ |
| `\\server\share` | Relative | OS-absolute UNC ✅ | OS-absolute UNC ✅ |
| `relative/path` | Relative | Relative | Relative |

---

## Examples

### Cross-Platform CLI Tool

```go
// filesystem_windows.go
//go:build windows

package mycli

import "github.com/absfs/osfs"

func NewFS(drive string) absfs.FileSystem {
	if drive == "" {
		drive = "C:"
	}
	return osfs.NewWindowsDriveMapper(osfs.NewFS(), drive)
}
```

```go
// filesystem_unix.go
//go:build !windows

package mycli

import "github.com/absfs/osfs"

func NewFS(drive string) absfs.FileSystem {
	return osfs.NewFS()
}
```

```go
// main.go
package main

import (
	"flag"
	"mycli"
)

func main() {
	drive := flag.String("drive", "", "Windows drive letter (e.g., 'D:')")
	flag.Parse()

	fs := mycli.NewFS(*drive)

	// Works on all platforms
	fs.MkdirAll("/var/log/mycli", 0755)
	fs.Create("/etc/mycli/config.json")
}
```

### Virtual Filesystem (No Setup Needed)

```go
package main

import "github.com/absfs/memfs"

func main() {
	fs := memfs.NewFS()

	// Just works everywhere - no platform-specific code needed!
	fs.Create("/config/app.json")
	fs.MkdirAll("/var/log/app", 0755)
}
```

### Testing

```go
func TestMyApp(t *testing.T) {
	// Virtual filesystem for testing
	fs := memfs.NewFS()

	// Unix-style paths work on all platforms
	err := fs.MkdirAll("/test/data", 0755)
	if err != nil {
		t.Fatal(err)
	}

	f, err := fs.Create("/test/data/file.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// Test passes on Windows, macOS, and Linux ✅
}
```

---

## Path Utilities

Use Go's `path/filepath` package for path operations:

```go
import "path/filepath"

// Platform-independent path joining
path := filepath.Join("/", "config", "app.json")
// Unix/macOS: /config/app.json
// Windows:    \config\app.json

// Clean paths
cleaned := filepath.Clean("/path/../to/file")
// Result: /to/file (or \to\file on Windows)

// Get directory
dir := filepath.Dir("/path/to/file.txt")
// Result: /path/to (or \path\to on Windows)

// Get filename
name := filepath.Base("/path/to/file.txt")
// Result: file.txt (all platforms)
```

---

## Summary

**For virtual filesystems:** Just use Unix-style `/paths` - they work everywhere automatically.

**For OS filesystems:** Use the cross-platform setup pattern with build tags. Create two tiny files (`filesystem_windows.go` and `filesystem_unix.go`), and your application code works identically on all platforms.

**The pattern gives you:**
- ✅ Write code once, runs everywhere
- ✅ Unix-style paths work on Windows
- ✅ Explicit drive selection when needed
- ✅ Zero overhead on Unix/macOS
- ✅ Standard Go build tag pattern

See [osfs WindowsDriveMapper](https://github.com/absfs/osfs) for implementation details.

For complete usage examples, see the [User Guide](USER_GUIDE.md).
