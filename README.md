# absfs - Abstract File System for go

[![Go Reference](https://pkg.go.dev/badge/github.com/absfs/absfs.svg)](https://pkg.go.dev/github.com/absfs/absfs)
[![Go Report Card](https://goreportcard.com/badge/github.com/absfs/absfs)](https://goreportcard.com/report/github.com/absfs/absfs)
[![Tests](https://github.com/absfs/absfs/workflows/Tests/badge.svg)](https://github.com/absfs/absfs/actions)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

`absfs` is a go package that defines an abstract filesystem interface.

The design goal of absfs is to support a system of composable filesystem implementations for various uses from testing to complex storage management.

Implementors can create stand alone filesystems that implement the `absfs.Filer` interface providing new components that can easily be added to existing compositions and data pipelines.

## Features

- 📦 **Composable** - Layer filesystems for complex behaviors
- 🧪 **Testable** - Mock filesystems for unit testing
- 🔧 **Extensible** - Easy to implement new filesystem types
- 🚀 **Performant** - Minimal overhead wrapper pattern
- 📚 **Well-documented** - Comprehensive docs and examples
- ✅ **Well-tested** - 89% code coverage
- 🌍 **Cross-platform** - Unix-style paths work on Windows, macOS, and Linux

## Install 

```bash
$ go get github.com/absfs/absfs
```

## The Interfaces 
`absfs` defines 4 interfaces.

1. `Filer` - The minimum set of methods that an implementation must define.
2. `FileSystem` - The complete set of Abstract FileSystem methods, including convenience functions.
3. `SymLinker` - Methods for implementing symbolic links.
4. `SymlinkFileSystem` - A FileSystem that supports symbolic links.

### Filer - Interface
The minimum set of methods that an implementation must define.

```go
type Filer interface {

    Mkdir(name string, perm os.FileMode) error
    OpenFile(name string, flag int, perm os.FileMode) File, error
    Remove(name string) error
    Rename(oldname, newname string) error
    Stat(name string) os.FileInfo, error
    Chmod(name string, mode os.FileMode) error
    Chtimes(name string, atime time.Time, mtime time.Time) error

}
```

### FileSystem - Interface
The complete set of Abstract FileSystem methods, including convenience functions.

```go
type FileSystem interface {

    // Filer interface
    OpenFile(name string, flag int, perm os.FileMode) (File, error)
    Mkdir(name string, perm os.FileMode) error
    Remove(name string) error
    Stat(name string) (os.FileInfo, error)
    Chmod(name string, mode os.FileMode) error
    Chtimes(name string, atime time.Time, mtime time.Time) error
    Chown(name string, uid, gid int) error

    Separator() uint8
    ListSeparator() uint8
    Chdir(dir string) error
    Getwd() (dir string, err error)
    TempDir() string
    Open(name string) (File, error)
    Create(name string) (File, error)
    MkdirAll(name string, perm os.FileMode) error
    RemoveAll(path string) (err error)
    Truncate(name string, size int64) error

}
```

### SymLinker - Interface
Additional methods for implementing symbolic links.

```go
type SymLinker interface {

    Lstat(fi1, fi2 os.FileInfo) bool
    Lchown(name string, uid, gid int) error
    Readlink(name string) (string, error)
    Symlink(oldname, newname string) error

}
```

### SymlinkFileSystem - Interface
A FileSystem that supports symbolic links.

```go
type SymlinkFileSystem interface {

    // Filer interface
    OpenFile(name string, flag int, perm os.FileMode) (File, error)
    Mkdir(name string, perm os.FileMode) error
    Remove(name string) error
    Stat(name string) (os.FileInfo, error)
    Chmod(name string, mode os.FileMode) error
    Chtimes(name string, atime time.Time, mtime time.Time) error
    Chown(name string, uid, gid int) error

    // FileSystem interface
    Separator() uint8
    ListSeparator() uint8
    Chdir(dir string) error
    Getwd() (dir string, err error)
    TempDir() string
    Open(name string) (File, error)
    Create(name string) (File, error)
    MkdirAll(name string, perm os.FileMode) error
    RemoveAll(path string) (err error)
    Truncate(name string, size int64) error

    // SymLinker interface
    Lstat(fi1, fi2 os.FileInfo) bool
    Lchown(name string, uid, gid int) error
    Readlink(name string) (string, error)
    Symlink(oldname, newname string) error
}
```

## File Systems

---

### [OsFs](https://github.com/absfs/osfs) 
An `absfs.FileSystem` implementation that wraps the equivalent `os` standard library filesystem functions. Using `osfs` is essentially identical to using the `os` package except that it implements a per object current working directory. 

### [MemFs](https://github.com/absfs/memfs)
The `memfs` filesystem implements a memory only filesystem.

### [NilFs](https://github.com/absfs/nilfs)
The Nil FileSystem is a no-op implementation. Methods have no effect and most return no errors. The `Read` and `ReadAt` methods always return `io.EOF` to avoid infinite loops.

### [PTFS](https://github.com/absfs/ptfs)
The Pass Through FileSystem (`ptfs`) takes any object that implements `absfs.FileSystem` and passes all methods through unaltered to that object.

### [ROFS](https://github.com/absfs/rofs)
The `rofs` is a read-only filesystem that can wrap any other filesystem implementation and make it read only.

### [BaseFS](https://github.com/absfs/basefs)
The `basefs` filesystem takes a `absfs.FileSystem` and an absolute path within that filesystem, and provides a `absfs.FileSystem` interface that is rooted at that path. Relative paths are handled correctly such that it is not possible to 'escape' the base path constraint (for example by using a path like "/.."). In addition errors are rewritten to accurately obscure the base path. 

### [HttpFS](https://github.com/absfs/httpfs)
A filesystem that can convert a `absfs.FileSystem` to a `http.FileSystem` interface for use with `http.FileServer`.

### [CorFS](https://github.com/absfs/corfs) (cache on read fs)
An `absfs.Filer` that wraps two `absfs.Filer`s when data is read if it doesn't exist in the second tier filesytem the filer copies data from the underlying file systems into the second tier filesystem causing it to function like a cache. 

### [CowFS](https://github.com/absfs/cowfs) (copy on write fs)
An `absfs.Filer` that wraps two `absfs.Filer`s the first of which may be a read only filesystem. Any attempt to modify or write data to a CowFS causes the Filer to write those changes to the secondary fs leaving the underlying filesystem unmodified.

### [BoltFS](https://github.com/absfs/boltfs)
An `absfs.Filer` that provides and absfs.Filer interface for the popular embeddable key/value store called bolt.

### [S3FS](https://github.com/absfs/s3fs)
An `absfs.Filer` that provides and absfs. Filer interface to any S3 compatible object storage API.

### [SftpFs](https://github.com/absfs/SftpFs)
A filesystem which reads and writes securely between computers across networks using the SFTP interface.

---

## Implementing a FileSystem
AbsFs compatible filesystems can and should be implemented in their own package and repo.
You may want to start with `nilfs`, `ptfs`, or `osfs` as a template. It is not necessary to implement all FileSystem methods on your custom FileSystem since you can use the `ExtendFiler` function to convert a Filer implementation to a full `FileSystem` implementation. 

#### Step 1 - Filer the minimum FileSystem interface
These methods are `OpenFile`, `Mkdir`, `Remove`, `Stat`, `Chmod`, `Chtimes`, `Chown`

```go
// template for a Filer implementation
type MyFiler struct {
    // ... 
}

func New() *MyFiler {
    return &MyFiler{}
}

func (fs *MyFiler) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
    // ...
}

func (fs *MyFiler) Mkdir(name string, perm os.FileMode) error  {
    // ...
}

func (fs *MyFiler) Remove(name string) error  {
    // ...
}

func (fs *MyFiler) Stat(name string) (os.FileInfo, error)  {
    // ...
}

func (fs *MyFiler) Chmod(name string, mode os.FileMode) error  {
    // ...
}

func (fs *MyFiler) Chtimes(name string, atime time.Time, mtime time.Time) error  {
    // ...
}

func (fs *MyFiler) Chown(name string, uid, gid int) error  {
    // ...
}


```

Optionally you can also implement any of the FileSystem interface methods if you need for performance or other reasons. The object returned by `ExtendFiler` only adds the missing methods `FileSystem` methods. The FileSystem interface adds the following additional methods: `Separator`, `ListSeparator`, `Chdir`, `Getwd`, `TempDir`, `Open`, `Create`, `MkdirAll` , `RemoveAll`, `Truncate`.

#### Step 2 - Extend A Filer to make a FileSystem
After implementing any of the additional FileSystem methods create a `NewFS` function that uses ExtendFiler to add the missing methods. 

```go
package myfs

import "github.com/absfs/absfs"

type MyFiler struct {
    // ...
}

// MyFiler implements the absfs.Filer interface

func NewFS() absfs.FileSystem {
    return absfs.ExtendFiler(&MyFiler{})
}

```

The implementation provided by ExtendFiler will first check for an existing method of the same signature on the underlying Filer.  If found it will call that method, if not it provides a default implementation.

An extended Filer implements the FileSystem interface as follows. If the FileSystem method is one of the convenience functions like `Open`, `Create`, or `MkdirAll` the default implementation simply uses the Filer methods (i.e. `OpenFile` and `Mkdir`) to implement the convenience function on top of the Filer interface.  If the missing methods is one of  `Separator`, `ListSeparator`, or `TempDir`, then the local operating system values are returned. Path navigation as provided by `Chdir`, and `Getwd` are provided as a complete path management implementation that resolves both absolute and relative paths much the same way as the `os` package. The extended filer will resolve paths into absolute paths and maintains a unique current working directory for each FileSystem interface object. A Filer is not required to implement relative paths.

## Thread Safety

⚠️ **FileSystem instances are NOT goroutine-safe by default.**

Each `FileSystem` object created by `ExtendFiler` maintains its own current working directory (`cwd`) state, which can be modified by `Chdir`. Concurrent access from multiple goroutines can cause race conditions.

### Safe Usage Patterns

1. **One FileSystem per goroutine** (Recommended)
   ```go
   func worker(id int) {
       fs := absfs.ExtendFiler(filer) // Each goroutine gets its own instance
       fs.Chdir(fmt.Sprintf("/worker%d", id))
       // ... do work ...
   }
   ```

2. **Use absolute paths only**
   ```go
   // Shared filesystem is safe if you only use absolute paths
   sharedFS := absfs.ExtendFiler(filer)
   // Safe: no dependency on cwd
   sharedFS.Open("/absolute/path/to/file.txt")
   ```

3. **External synchronization**
   ```go
   var mu sync.Mutex
   sharedFS := absfs.ExtendFiler(filer)

   mu.Lock()
   sharedFS.Chdir("/some/directory")
   file, _ := sharedFS.Open("relative/file.txt")
   mu.Unlock()
   ```

See [SECURITY.md](SECURITY.md) for more details.

## Path Handling

The `absfs` package provides consistent path semantics across all platforms. Paths starting with `/` or `\` are treated as **virtual absolute paths**, allowing Unix-style paths to work on Windows while maintaining support for OS-native paths.

```go
// These work identically on Unix, macOS, AND Windows:
fs := absfs.ExtendFiler(virtualFS)
fs.Create("/config/app.json")
fs.MkdirAll("/var/log/app", 0755)
```

**Important for Windows users**: Paths like `/config/file.txt` are treated as virtual-absolute for portability. For true OS operations on Windows, use drive letters (`C:\path`) or UNC paths (`\\server\share`).

See [PATH_HANDLING.md](PATH_HANDLING.md) for comprehensive cross-platform path behavior documentation.

## Documentation

- [Path Handling Guide](PATH_HANDLING.md) - Cross-platform path semantics
- [Architecture Guide](ARCHITECTURE.md) - Design patterns and internals
- [Security Policy](SECURITY.md) - Security considerations and best practices
- [Changelog](CHANGELOG.md) - Version history and changes
- [GoDoc](https://pkg.go.dev/github.com/absfs/absfs) - API documentation

## Contributing

We strongly encourage contributions, please fork and submit Pull Requests, and publish any FileSystems you implement.

New FileSystem types do not need to be added to this repo, but we'd be happy to link to yours so please open an issue or better yet add a Pull Request with the updated Readme.

## LICENSE

This project is governed by the MIT License. See [LICENSE](https://github.com/absfs/absfs/blob/master/LICENSE)

