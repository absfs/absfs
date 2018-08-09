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

type Filer interface {

	// OpenFile opens a file using the given flags and the given mode.
	OpenFile(name string, flag int, perm os.FileMode) (File, error)

	// Mkdir creates a directory in the filesystem, return an error if any
	// happens.
	Mkdir(name string, perm os.FileMode) error

	// Remove removes a file identified by name, returning an error, if any
	// happens.
	Remove(name string) error

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
func ExtendFiler(filer Filer) FileSystem {
	return &fs{"/", filer}
}

type fs struct {
	cwd   string
	filer Filer
}

func (fs *fs) OpenFile(name string, flag int, perm os.FileMode) (f File, err error) {
	if !filepath.IsAbs(name) {
		if _, ok := fs.filer.(dirnavigator); !ok {
			name = filepath.Clean(filepath.Join(fs.cwd, name))
		}
	}
	return fs.filer.OpenFile(name, flag, perm)
}

func (fs *fs) Mkdir(name string, perm os.FileMode) error {
	if !filepath.IsAbs(name) {
		if _, ok := fs.filer.(dirnavigator); !ok {
			name = filepath.Clean(filepath.Join(fs.cwd, name))
		}
	}
	return fs.filer.Mkdir(name, perm)
}

func (fs *fs) Remove(name string) error {
	if !filepath.IsAbs(name) {
		if _, ok := fs.filer.(dirnavigator); !ok {
			name = filepath.Clean(filepath.Join(fs.cwd, name))
		}
	}
	return fs.filer.Remove(name)
}

func (fs *fs) Stat(name string) (os.FileInfo, error) {
	if !filepath.IsAbs(name) {
		if _, ok := fs.filer.(dirnavigator); !ok {
			name = filepath.Clean(filepath.Join(fs.cwd, name))
		}
	}
	return fs.filer.Stat(name)
}

func (fs *fs) Chmod(name string, mode os.FileMode) error {
	if !filepath.IsAbs(name) {
		if _, ok := fs.filer.(dirnavigator); !ok {
			name = filepath.Clean(filepath.Join(fs.cwd, name))
		}
	}
	return fs.filer.Chmod(name, mode)
}

func (fs *fs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	if !filepath.IsAbs(name) {
		if _, ok := fs.filer.(dirnavigator); !ok {
			name = filepath.Clean(filepath.Join(fs.cwd, name))
		}
	}

	return fs.filer.Chtimes(name, atime, mtime)
}

func (fs *fs) Chown(name string, uid, gid int) error {
	if !filepath.IsAbs(name) {
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
	if !filepath.IsAbs(name) {
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
	if !filepath.IsAbs(name) {
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
	if !filepath.IsAbs(name) {
		if _, ok := fs.filer.(dirnavigator); !ok {
			name = filepath.Clean(filepath.Join(fs.cwd, name))
		}
	}

	path := string(fs.Separator())
	for _, p := range strings.Split(name, string(fs.Separator())) {
		if p == "" {
			continue
		}
		path = filepath.Join(path, p)
		if path == "/" {
			continue
		}
		fs.Mkdir(path, perm)
	}

	return nil
}

func (fs *fs) removeAll(path string) (err error) {
	// open the file
	var f File
	f, err = fs.Open(path) // .filer.Open(path)
	if err != nil {
		return err
	}

	// defer close with error checking, and remove the file after closing.
	defer func() {
		if err != nil {
			f.Close()
			// return err
		}
		closer := f.Close()
		err = fs.filer.Remove(path)
		if err == nil {
			err = closer
		}
		// return err
		return
	}()

	// get FileInfo
	info, err := f.Stat()
	if err != nil {
		return err
	}

	// if it's not a directory remove it and return
	if !info.IsDir() {
		return fs.Remove(path)
	}

	// get and loop through each directory entry calling remove all recursively.
	err = nil
	var names []string
	for err != io.EOF {
		names, err = f.Readdirnames(512)
		if err != nil && len(names) == 0 {
			if err != io.EOF {
				return err
			}
		}
		for _, name := range names {
			if name == "." || name == ".." {
				continue
			}
			err := fs.removeAll(filepath.Join(path, name))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (fs *fs) RemoveAll(name string) (err error) {

	if filer, ok := fs.filer.(remover); ok {
		return filer.RemoveAll(name)
	}
	if !filepath.IsAbs(name) {
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
	if !filepath.IsAbs(name) {
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
