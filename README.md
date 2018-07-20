# absfs - Abstract File System for go
`absfs` is a go package that defines an abstract filesystem interface.

The design goal of absfs is to support a system of composable filesystem implementations for various uses from testing to complex storage management.

Implementors can create stand alone filesystems that implement the abfs.Filer interface providing new components that can easily be add to existing compositions, and data pipelines.

absfs will provide a testing suite to help implementers produce packages that function and error consistently with all other abstract filesystems implementations. 


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

## AbsFS Provided filesystems

---

### NilFs 
An `absfs.Filer` implementation that does nothing and returns no errors. Maybe be useful as a starting place for new file system implementations.

### PassThroughFs 
An `absfs.Filer` implementation that wraps another `absfs.Filer` passing all operation through unaltered to the underlying filesystem. Also maybe be useful as a starting place for new file system implementations.

### [OsFs](https://github.com/absfs/osfs) 
An `absfs.FileSystem` implementation that wraps the `os` standard library package file handling functions.

### MemFs
An `absfs.Filer` implementation that provides an in memory ephemeral filesystem. 

### ReadOnlyFs
An `absfs.Filer` that wraps any other `absfs.Filer` and intercepts any attempt to write data and responds with the appropriate read only permissions errors.

### BaseFS
An `absfs.Filer` that wraps any other `absfs.Filer` and forces all path operations to be constrained to folder of the underlying filesystem. 

### HttpFS
A `http.FileSystem` that wraps any `absfs.Filer`.

### CorFS (cache on read fs)
An `absfs.Filer` that wraps two `absfs.Filer`s when data is read if it doesn't exist in the second tier filesytem the filer copies data from the underlying file systems into the second tier filesystem causing it to function like a cache. 

### CowFS (copy on write fs)
An `absfs.Filer` that wraps two `absfs.Filer`s the first of which may be a read only filesystem. Any attempt to modify or write data to a CowFS causes the Filer to write those changes to the secondary fs leaving the underlying filesystem unmodified.

### BoltFS
An `absfs.Filer` that provides and absfs.Filer interface for the popular embeddable key/value store called bolt.

### S3FS
An `absfs.Filer` that provides and absfs. Filer interface to any S3 compatible object storage API.

### SftpFS
A filesystem which reads and writes securely between computers across networks using the SFTP interface.

---

## Implementing a FileSystem
AbsFs compatible filesystems can and should be implemented in their own package and repo.
You may want to start with NilFs, PassThroughFs, or OsFs as a template if you plan to implement all of the FileSystem interface methods. However, it is not necessary to implement all FileSystem methods on your custom FileSystem since you can use the ExtendFiler function to convert a Filer implementation to a full FileSystem implementation. This means you only need to implement the 7 Filer methods to get all AbsFs features.

#### Step 1 - Implemented The 7 Filer interface methods
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

Optionally you can also implement any of the FileSystem interface methods if you need custom behavior.  These methods will be called instead of the ExtendFiler added methods (more info on how this is done below). The FileSystem interface adds the following additional methods: `Separator`, `ListSeparator`, `Chdir`, `Getwd`, `TempDir`, `Open`, `Create`, `MkdirAll` , `RemoveAll`, `Truncate`.

#### Step 2 - Use ExtendFiler to create a complete FileSystem implementation from a Filer
After optionally implementing any of the additional FileSystem methods create a `NewFS` function that uses ExtendFiler to add the missing methods. 

```go
package myfs

import "github.com/absfs/absfs"

type MyFiler struct {
    // ...
}

// ...

func NewFS() absfs.FileSystem {
    return absfs.ExtendFiler(&MyFiler{})
}

```

The implementation provided by ExtendFiler will first check for an existing method of the same signature on the underlying Filer.  If found it will simply call that method, and return the results, however if no method is found then a generic implementation is provided. If the FileSystem method is one of the convenience functions like `Open`, `Create`, or `MkdirAll` the default implementation simply uses the Filer methods (i.e. `OpenFile` and `Mkdir`) to implement the convenience function on top of the Filer interface.  If the missing methods is one of  `Separator`, `ListSeparator`, or `TempDir`, then the local operating system values are returned. Path navigation as provided by `Chdir`, and `Getwd` are provided as a complete path management implementation layered on top of the Filer so that the filer only needs to handle absolute paths. 

## Contributing

We strongly encourage contributions, please fork and submit Pull Requests.

New FileSystem types do not need to be added to this repo, but if they are good quality implementations of useful new FileSystems we will happily add a link and description to this Readme. 

## LICENSE

This project is governed by the MIT License. See [LICENSE](https://github.com/absfs/absfs/blob/master/LICENSE)

