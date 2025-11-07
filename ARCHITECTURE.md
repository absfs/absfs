# absfs Architecture

## Design Philosophy

absfs is designed around the principle of **composability**. Rather than providing a single monolithic filesystem implementation, absfs defines interfaces that allow different filesystem implementations to be composed together to create complex storage solutions.

## Core Concepts

### 1. Minimal Interface (Filer)

The `Filer` interface defines the absolute minimum set of methods required for a basic filesystem:

```go
type Filer interface {
    OpenFile(name string, flag int, perm os.FileMode) (File, error)
    Mkdir(name string, perm os.FileMode) error
    Remove(name string) error
    Rename(oldpath, newpath string) error
    Stat(name string) (os.FileInfo, error)
    Chmod(name string, mode os.FileMode) error
    Chtimes(name string, atime time.Time, mtime time.Time) error
    Chown(name string, uid, gid int) error
}
```

**Design Decision**: By keeping the required interface small, it's easier for implementers to create new filesystem types.

### 2. Extended Interface (FileSystem)

The `FileSystem` interface extends `Filer` with convenience methods:

- `Open`, `Create` - Simplified file opening
- `MkdirAll`, `RemoveAll` - Recursive operations
- `Separator`, `ListSeparator`, `TempDir` - Platform-specific values
- `Chdir`, `Getwd` - Directory navigation

**Design Decision**: These methods are provided automatically via `ExtendFiler`, so implementers don't have to write them unless they want optimized versions.

### 3. Optional Interfaces

absfs uses optional interfaces for advanced features. Implementations can optionally implement:

- `opener` - Custom Open logic
- `creator` - Custom Create logic
- `mkaller` - Optimized MkdirAll
- `remover` - Optimized RemoveAll
- `separator`, `listseparator` - Platform separators
- `dirnavigator` - Custom Chdir/Getwd
- `temper` - Custom TempDir
- `truncater` - Custom Truncate

**Design Decision**: Type assertions (`interface.(type)`) check if optional interfaces are implemented. If yes, the custom implementation is used; otherwise, a default implementation is provided.

## Implementation Pattern

```
┌─────────────┐
│ FileSystem  │ (Extended interface with all methods)
└──────┬──────┘
       │
   ExtendFiler() - Wraps Filer and adds missing methods
       │
┌──────▼──────┐
│    Filer    │ (Minimal interface - user implements this)
└─────────────┘
```

### How ExtendFiler Works

```go
func ExtendFiler(filer Filer) FileSystem {
    return &fs{"/", filer} // Wraps filer with default implementations
}
```

The `fs` struct:
- Stores current working directory (cwd)
- Wraps the user's `Filer` implementation
- Provides default implementations for FileSystem methods
- Converts relative paths to absolute paths using cwd
- Uses type assertions to call optional interface methods if available

## Path Handling

### Absolute vs Relative Paths

```go
func (fs *fs) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
    // Check if path is relative
    if !filepath.IsAbs(name) {
        // Check if filer implements directory navigation
        if _, ok := fs.filer.(dirnavigator); !ok {
            // Convert to absolute using cwd
            name = filepath.Clean(filepath.Join(fs.cwd, name))
        }
    }
    return fs.filer.OpenFile(name, flag, perm)
}
```

**Design Decision**:
- If the filer implements `dirnavigator`, it handles relative paths itself
- Otherwise, the wrapper converts relative → absolute using the wrapper's cwd
- This allows simple Filer implementations to only handle absolute paths

## File Interface Hierarchy

```
File (Full interface)
  │
  ├─ ReadAt, WriteAt, WriteString, Truncate, Readdirnames
  │
Seekable
  │
  ├─ Seek, Read, Write, Close, Stat, Readdir, Sync
  │
UnSeekable
  │
  └─ Read, Write, Close, Stat, Readdir, Sync, Name
```

Similar to Filer→FileSystem, there's `ExtendSeekable(Seekable) File` which adds:
- ReadAt/WriteAt (implemented via Seek + Read/Write)
- Truncate (implemented via Seek + Write of zeros)
- Readdirnames (implemented via Readdir)

## Composition Patterns

### Layering Example

```
Application
    │
    ▼
BaseFS (constrains to /app/data)
    │
    ▼
CowFS (writes go to overlay)
    │
    ├─▶ MemFS (overlay - fast writes)
    │
    └─▶ ROFS (base - read-only)
            │
            ▼
        OsFS (actual disk)
```

Each layer wraps the one below, adding functionality.

## Key Design Patterns

### 1. Interface Segregation
- Small, focused interfaces
- Optional extensions via type assertions
- No bloated monolithic interfaces

### 2. Composition Over Inheritance
- Filesystems wrap other filesystems
- Behavior added by wrapping, not subclassing

### 3. Fail-Safe Defaults
- ExtendFiler provides working defaults
- Implementations can optimize by implementing optional interfaces
- Relative path handling works automatically

### 4. Transparency
- File methods pass through to underlying implementation
- Wrapper adds minimal overhead
- Type assertions happen once per method call

## Testing Strategy

### Mock Implementations
Tests use mock Filers to verify the wrapper logic works correctly:
- `mockFiler` - Basic implementation
- `mockFilerWithOptionals` - Implements all optional interfaces
- `testSeekable` - For testing File adapters

### Coverage Goals
- Core interfaces: 100%
- Wrapper logic: 90%+
- Optional interface paths: Both branches tested

## Performance Considerations

### Path Operations
- `filepath.Clean` and `filepath.Join` called for relative paths
- Absolute paths pass through with minimal overhead
- Per-instance cwd allows isolation

### Type Assertions
- Checked once per method call
- Negligible overhead
- Enables zero-cost abstraction when optional interfaces are implemented

### Memory
- Each FileSystem wrapper: ~16 bytes (cwd string + filer pointer)
- Minimal allocation for path operations

## Thread Safety

⚠️ **FileSystem instances are NOT thread-safe by default**

- Each instance has mutable `cwd` state
- Concurrent `Chdir` calls cause races
- Solutions:
  1. One FileSystem per goroutine
  2. Use absolute paths only
  3. External synchronization
  4. Implement thread-safe wrapper

## Extension Points

To create a new filesystem:

1. **Implement Filer** (8 methods minimum)
2. **Optionally implement optional interfaces** for optimization
3. **Use ExtendFiler** to get full FileSystem
4. **Document thread safety** and limitations

Example:
```go
type MyFS struct { /* ... */ }

func (m *MyFS) OpenFile(...) { /* ... */ }
// ... implement other Filer methods ...

func NewMyFS() absfs.FileSystem {
    return absfs.ExtendFiler(&MyFS{})
}
```

## Future Considerations

- Context support for cancellation
- Metrics/observability hooks
- Pluggable path normalization
- Advanced caching strategies
- Distributed filesystem support
