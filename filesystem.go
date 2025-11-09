package absfs

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrNotImplemented = errors.New("not implemented")

type ReadOnlyFiler interface {
	Open(name string) (io.ReadCloser, error)
}

type WriteOnlyFiler interface {
	Open(name string) (io.WriteCloser, error)
}

type Filer interface {

	// OpenFile opens a file using the given flags and the given mode.
	OpenFile(name string, flag int, perm os.FileMode) (File, error)

	// Mkdir creates a directory in the filesystem, return an error if any
	// happens.
	Mkdir(name string, perm os.FileMode) error

	// Remove removes a file identified by name, returning an error, if any
	// happens.
	Remove(name string) error

	// Rename renames (moves) oldpath to newpath. If newpath already exists and
	// is not a directory, Rename replaces it. OS-specific restrictions may apply
	// when oldpath and newpath are in different directories. If there is an
	// error, it will be of type *LinkError.
	Rename(oldpath, newpath string) error

	// Stat returns the FileInfo structure describing file. If there is an error,
	// it will be of type *PathError.
	Stat(name string) (os.FileInfo, error)

	//Chmod changes the mode of the named file to mode.
	Chmod(name string, mode os.FileMode) error

	//Chtimes changes the access and modification times of the named file
	Chtimes(name string, atime time.Time, mtime time.Time) error

	//Chown changes the owner and group ids of the named file
	Chown(name string, uid, gid int) error
}

type FileSystem interface {
	Filer

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

type SymLinker interface {
	// SameFile(fi1, fi2 os.FileInfo) bool

	// Lstat returns a FileInfo describing the named file. If the file is a
	// symbolic link, the returned FileInfo describes the symbolic link. Lstat
	// makes no attempt to follow the link. If there is an error, it will be of type *PathError.
	Lstat(name string) (os.FileInfo, error)

	// Lchown changes the numeric uid and gid of the named file. If the file is a
	// symbolic link, it changes the uid and gid of the link itself. If there is
	// an error, it will be of type *PathError.
	//
	// On Windows, it always returns the syscall.EWINDOWS error, wrapped in
	// *PathError.
	Lchown(name string, uid, gid int) error

	// Readlink returns the destination of the named symbolic link. If there is an
	// error, it will be of type *PathError.
	Readlink(name string) (string, error)

	// Symlink creates newname as a symbolic link to oldname. If there is an
	// error, it will be of type *LinkError.
	Symlink(oldname, newname string) error
}

type SymlinkFileSystem interface {
	FileSystem
	SymLinker
}

// ExtendFiler adds the FileSystem convenience functions to any Filer implementation.
//
// Path Semantics:
//
// The returned FileSystem treats paths starting with '/' or '\' as absolute within
// the virtual filesystem, even on Windows where these aren't truly OS-absolute
// (which require drive letters like "C:\"). This design enables virtual filesystems
// (mocks, in-memory, archives) to use Unix-style paths consistently across all
// platforms while still supporting true OS-absolute paths.
//
// Examples:
//   - "/config/app.json"       → virtual-absolute on all platforms
//   - "C:\Windows\file.txt"    → OS-absolute on Windows (also virtual-absolute)
//   - "\\server\share\file"    → OS-absolute UNC on Windows (also virtual-absolute)
//   - "relative/path"          → relative on all platforms
//
// For most use cases with virtual filesystems, simply use Unix-style absolute
// paths ("/path/to/file") and they will work correctly across all platforms.
// For OS filesystem wrappers, use platform-native paths for best results.
//
// See PATH_HANDLING.md for detailed cross-platform behavior documentation.
func ExtendFiler(filer Filer) FileSystem {
	return &fs{string(filepath.Separator), filer}
}

type fs struct {
	cwd   string
	filer Filer
}

// isVirtualAbs checks if a path should be treated as absolute in the virtual filesystem.
// On Unix, this matches filepath.IsAbs. On Windows, we also treat paths starting with
// '/' or '\' as absolute, even though they're not OS-absolute (lack drive letter).
func isVirtualAbs(path string) bool {
	if filepath.IsAbs(path) {
		return true
	}
	// Treat paths starting with separator as absolute in virtual filesystem
	if len(path) > 0 && (path[0] == '/' || path[0] == '\\') {
		return true
	}
	return false
}

func (fs *fs) OpenFile(name string, flag int, perm os.FileMode) (f File, err error) {
	if !isVirtualAbs(name) {
		if _, ok := fs.filer.(dirnavigator); !ok {
			name = filepath.Clean(filepath.Join(fs.cwd, name))
		}
	}
	return fs.filer.OpenFile(name, flag, perm)
}

func (fs *fs) Mkdir(name string, perm os.FileMode) error {
	if !isVirtualAbs(name) {
		if _, ok := fs.filer.(dirnavigator); !ok {
			name = filepath.Clean(filepath.Join(fs.cwd, name))
		}
	}
	return fs.filer.Mkdir(name, perm)
}

func (fs *fs) Remove(name string) error {
	if !isVirtualAbs(name) {
		if _, ok := fs.filer.(dirnavigator); !ok {
			name = filepath.Clean(filepath.Join(fs.cwd, name))
		}
	}
	return fs.filer.Remove(name)
}

func (fs *fs) Rename(oldpath, newpath string) error {
	if !isVirtualAbs(oldpath) {
		if _, ok := fs.filer.(dirnavigator); !ok {
			oldpath = filepath.Clean(filepath.Join(fs.cwd, oldpath))
		}
	}
	if !isVirtualAbs(newpath) {
		if _, ok := fs.filer.(dirnavigator); !ok {
			newpath = filepath.Clean(filepath.Join(fs.cwd, newpath))
		}
	}

	return fs.filer.Rename(oldpath, newpath)
}

func (fs *fs) Stat(name string) (os.FileInfo, error) {
	if !isVirtualAbs(name) {
		if _, ok := fs.filer.(dirnavigator); !ok {
			name = filepath.Clean(filepath.Join(fs.cwd, name))
		}
	}
	return fs.filer.Stat(name)
}

func (fs *fs) Chmod(name string, mode os.FileMode) error {
	if !isVirtualAbs(name) {
		if _, ok := fs.filer.(dirnavigator); !ok {
			name = filepath.Clean(filepath.Join(fs.cwd, name))
		}
	}
	return fs.filer.Chmod(name, mode)
}

func (fs *fs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	if !isVirtualAbs(name) {
		if _, ok := fs.filer.(dirnavigator); !ok {
			name = filepath.Clean(filepath.Join(fs.cwd, name))
		}
	}

	return fs.filer.Chtimes(name, atime, mtime)
}

func (fs *fs) Chown(name string, uid, gid int) error {
	if !isVirtualAbs(name) {
		if _, ok := fs.filer.(dirnavigator); !ok {
			name = filepath.Clean(filepath.Join(fs.cwd, name))
		}
	}
	return fs.filer.Chown(name, uid, gid)
}

func (fs *fs) Separator() uint8 {
	if filer, ok := fs.filer.(separator); ok {
		return filer.Separator()
	}
	return filepath.Separator
}

func (fs *fs) ListSeparator() uint8 {
	if filer, ok := fs.filer.(listseparator); ok {
		return filer.ListSeparator()
	}
	return filepath.ListSeparator
}

func (fs *fs) Chdir(dir string) error {
	if filer, ok := fs.filer.(dirnavigator); ok {
		return filer.Chdir(dir)
	}
	f, err := fs.Open(dir)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return &os.PathError{Op: "chdir", Path: dir, Err: errors.New("not a directory")}
	}
	fs.cwd = filepath.Clean(dir)
	return nil
}

func (fs *fs) Getwd() (dir string, err error) {
	if filer, ok := fs.filer.(dirnavigator); ok {
		return filer.Getwd()
	}
	return fs.cwd, nil
}

func (fs *fs) TempDir() string {
	if filer, ok := fs.filer.(temper); ok {
		return filer.TempDir()
	}

	return os.TempDir()
}

func (fs *fs) Open(name string) (File, error) {
	if filer, ok := fs.filer.(opener); ok {
		return filer.Open(name)
	}
	if !isVirtualAbs(name) {
		if _, ok := fs.filer.(dirnavigator); !ok {
			name = filepath.Clean(filepath.Join(fs.cwd, name))
		}
	}
	return fs.filer.OpenFile(name, os.O_RDONLY, 0)
}

func (fs *fs) Create(name string) (File, error) {
	if filer, ok := fs.filer.(creator); ok {
		return filer.Create(name)
	}
	if !isVirtualAbs(name) {
		if _, ok := fs.filer.(dirnavigator); !ok {
			name = filepath.Clean(filepath.Join(fs.cwd, name))
		}
	}
	return fs.filer.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
}

func (fs *fs) MkdirAll(name string, perm os.FileMode) error {
	if filer, ok := fs.filer.(mkaller); ok {
		return filer.MkdirAll(name, perm)
	}
	if !isVirtualAbs(name) {
		if _, ok := fs.filer.(dirnavigator); !ok {
			name = filepath.Clean(filepath.Join(fs.cwd, name))
		}
	}

	// Normalize path separators to platform-specific separator before splitting
	// This ensures /a/b/c works correctly on Windows (converts to \a\b\c)
	name = filepath.Clean(name)

	path := string(fs.Separator())
	for _, p := range strings.Split(name, string(fs.Separator())) {
		if p == "" {
			continue
		}
		path = filepath.Join(path, p)
		if path == string(filepath.Separator) {
			continue
		}
		fs.Mkdir(path, perm)
	}

	return nil
}

func (fs *fs) removeAll(path string) error {
	// open the file to check if it's a directory
	f, err := fs.Open(path)
	if err != nil {
		return err
	}

	// get FileInfo
	info, err := f.Stat()
	closeErr := f.Close()
	if err != nil {
		return err
	}

	// if it's not a directory, just remove it
	if !info.IsDir() {
		return fs.Remove(path)
	}

	// For directories, we need to recursively remove contents
	// Reopen to read directory entries
	f, err = fs.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Read all directory entries and remove them recursively
	for {
		names, err := f.Readdirnames(512)
		for _, name := range names {
			if name == "." || name == ".." {
				continue
			}
			if err := fs.removeAll(filepath.Join(path, name)); err != nil {
				return err
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	// Close the directory before removing it
	if err := f.Close(); err != nil && closeErr == nil {
		closeErr = err
	}

	// Finally, remove the directory itself
	if err := fs.filer.Remove(path); err != nil {
		return err
	}

	return closeErr
}

func (fs *fs) RemoveAll(name string) (err error) {

	if filer, ok := fs.filer.(remover); ok {
		return filer.RemoveAll(name)
	}
	if !isVirtualAbs(name) {
		if _, ok := fs.filer.(dirnavigator); !ok {
			name = filepath.Clean(filepath.Join(fs.cwd, name))
		}
	}
	return fs.removeAll(name)
}

func (fs *fs) Truncate(name string, size int64) error {
	if filer, ok := fs.filer.(truncater); ok {
		return filer.Truncate(name, size)
	}
	if !isVirtualAbs(name) {
		if _, ok := fs.filer.(dirnavigator); !ok {
			name = filepath.Clean(filepath.Join(fs.cwd, name))
		}
	}

	f, err := fs.filer.OpenFile(name, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	return f.Close()
}

type FastWalkFunc func(string, os.FileMode) error

// interfaces for easy method typing

type opener interface {
	Open(name string) (File, error)
}

type creator interface {
	Create(name string) (File, error)
}

type mkaller interface {
	MkdirAll(name string, perm os.FileMode) error
}

type remover interface {
	RemoveAll(path string) (err error)
}

type separator interface {
	Separator() uint8
}

type listseparator interface {
	ListSeparator() uint8
}

type dirnavigator interface {
	Chdir(dir string) error
	Getwd() (dir string, err error)
}

type temper interface {
	TempDir() string
}

type truncater interface {
	Truncate(name string, size int64) error
}
