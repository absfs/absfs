package absfs

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// mockFiler is a minimal mock implementation of Filer for testing
type mockFiler struct {
	files map[string]*mockFile
}

func newMockFiler() *mockFiler {
	return &mockFiler{
		files: make(map[string]*mockFile),
	}
}

func (m *mockFiler) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	name = filepath.Clean(name)
	if flag&os.O_CREATE != 0 {
		if _, exists := m.files[name]; !exists {
			m.files[name] = &mockFile{
				name:    name,
				mode:    perm,
				content: []byte{},
				modTime: time.Now(),
			}
		}
	}

	f, exists := m.files[name]
	if !exists {
		return &InvalidFile{Path: name}, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
	}

	if flag&os.O_TRUNC != 0 {
		f.content = []byte{}
	}

	return &mockFileHandle{file: f, offset: 0}, nil
}

func (m *mockFiler) Mkdir(name string, perm os.FileMode) error {
	name = filepath.Clean(name)
	if _, exists := m.files[name]; exists {
		return &os.PathError{Op: "mkdir", Path: name, Err: os.ErrExist}
	}
	m.files[name] = &mockFile{
		name:  name,
		mode:  os.ModeDir | perm,
		isDir: true,
		modTime: time.Now(),
	}
	return nil
}

func (m *mockFiler) Remove(name string) error {
	name = filepath.Clean(name)
	if _, exists := m.files[name]; !exists {
		return &os.PathError{Op: "remove", Path: name, Err: os.ErrNotExist}
	}
	delete(m.files, name)
	return nil
}

func (m *mockFiler) Rename(oldpath, newpath string) error {
	oldpath = filepath.Clean(oldpath)
	newpath = filepath.Clean(newpath)
	f, exists := m.files[oldpath]
	if !exists {
		return &os.PathError{Op: "rename", Path: oldpath, Err: os.ErrNotExist}
	}
	delete(m.files, oldpath)
	f.name = newpath
	m.files[newpath] = f
	return nil
}

func (m *mockFiler) Stat(name string) (os.FileInfo, error) {
	name = filepath.Clean(name)
	f, exists := m.files[name]
	if !exists {
		return nil, &os.PathError{Op: "stat", Path: name, Err: os.ErrNotExist}
	}
	return f, nil
}

func (m *mockFiler) Chmod(name string, mode os.FileMode) error {
	name = filepath.Clean(name)
	f, exists := m.files[name]
	if !exists {
		return &os.PathError{Op: "chmod", Path: name, Err: os.ErrNotExist}
	}
	f.mode = mode
	return nil
}

func (m *mockFiler) Chtimes(name string, atime time.Time, mtime time.Time) error {
	name = filepath.Clean(name)
	f, exists := m.files[name]
	if !exists {
		return &os.PathError{Op: "chtimes", Path: name, Err: os.ErrNotExist}
	}
	f.modTime = mtime
	return nil
}

func (m *mockFiler) Chown(name string, uid, gid int) error {
	name = filepath.Clean(name)
	if _, exists := m.files[name]; !exists {
		return &os.PathError{Op: "chown", Path: name, Err: os.ErrNotExist}
	}
	return nil
}

// mockFile represents a file in the mock filesystem
type mockFile struct {
	name    string
	mode    os.FileMode
	isDir   bool
	content []byte
	modTime time.Time
	entries []os.FileInfo
}

func (f *mockFile) Name() string       { return filepath.Base(f.name) }
func (f *mockFile) Size() int64        { return int64(len(f.content)) }
func (f *mockFile) Mode() os.FileMode  { return f.mode }
func (f *mockFile) ModTime() time.Time { return f.modTime }
func (f *mockFile) IsDir() bool        { return f.isDir }
func (f *mockFile) Sys() interface{}   { return nil }

// mockFileHandle implements the File interface
type mockFileHandle struct {
	file   *mockFile
	offset int64
}

func (fh *mockFileHandle) Name() string { return fh.file.name }

func (fh *mockFileHandle) Read(b []byte) (int, error) {
	if fh.offset >= int64(len(fh.file.content)) {
		return 0, io.EOF
	}
	n := copy(b, fh.file.content[fh.offset:])
	fh.offset += int64(n)
	if n < len(b) {
		return n, io.EOF
	}
	return n, nil
}

func (fh *mockFileHandle) Write(b []byte) (int, error) {
	if fh.offset > int64(len(fh.file.content)) {
		padding := make([]byte, fh.offset-int64(len(fh.file.content)))
		fh.file.content = append(fh.file.content, padding...)
	}

	if fh.offset == int64(len(fh.file.content)) {
		fh.file.content = append(fh.file.content, b...)
	} else {
		end := fh.offset + int64(len(b))
		if end > int64(len(fh.file.content)) {
			fh.file.content = append(fh.file.content[:fh.offset], b...)
		} else {
			copy(fh.file.content[fh.offset:], b)
		}
	}

	fh.offset += int64(len(b))
	return len(b), nil
}

func (fh *mockFileHandle) Close() error { return nil }
func (fh *mockFileHandle) Sync() error  { return nil }

func (fh *mockFileHandle) Stat() (os.FileInfo, error) {
	return fh.file, nil
}

func (fh *mockFileHandle) Readdir(n int) ([]os.FileInfo, error) {
	if !fh.file.isDir {
		return nil, errors.New("not a directory")
	}
	return fh.file.entries, nil
}

func (fh *mockFileHandle) Seek(offset int64, whence int) (int64, error) {
	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = fh.offset + offset
	case io.SeekEnd:
		newOffset = int64(len(fh.file.content)) + offset
	default:
		return 0, errors.New("invalid whence")
	}

	if newOffset < 0 {
		return 0, errors.New("negative position")
	}

	fh.offset = newOffset
	return newOffset, nil
}

func (fh *mockFileHandle) ReadAt(b []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, errors.New("negative offset")
	}
	if off >= int64(len(fh.file.content)) {
		return 0, io.EOF
	}
	n = copy(b, fh.file.content[off:])
	if n < len(b) {
		err = io.EOF
	}
	return n, err
}

func (fh *mockFileHandle) WriteAt(b []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, errors.New("negative offset")
	}

	if off > int64(len(fh.file.content)) {
		padding := make([]byte, off-int64(len(fh.file.content)))
		fh.file.content = append(fh.file.content, padding...)
	}

	if off+int64(len(b)) > int64(len(fh.file.content)) {
		fh.file.content = append(fh.file.content[:off], b...)
	} else {
		copy(fh.file.content[off:], b)
	}

	return len(b), nil
}

func (fh *mockFileHandle) WriteString(s string) (n int, err error) {
	return fh.Write([]byte(s))
}

func (fh *mockFileHandle) Truncate(size int64) error {
	if size < 0 {
		return errors.New("negative size")
	}
	if size > int64(len(fh.file.content)) {
		padding := make([]byte, size-int64(len(fh.file.content)))
		fh.file.content = append(fh.file.content, padding...)
	} else {
		fh.file.content = fh.file.content[:size]
	}
	return nil
}

func (fh *mockFileHandle) Readdirnames(n int) (names []string, err error) {
	if !fh.file.isDir {
		return nil, errors.New("not a directory")
	}
	for _, entry := range fh.file.entries {
		names = append(names, entry.Name())
	}
	return names, nil
}

// Tests start here

func TestExtendFiler(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	if fs == nil {
		t.Fatal("ExtendFiler returned nil")
	}
}

func TestFileSystemCreate(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	f, err := fs.Create("/test.txt")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer f.Close()

	if f.Name() != "/test.txt" {
		t.Errorf("expected name /test.txt, got %s", f.Name())
	}
}

func TestFileSystemOpen(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	// Create a file first
	f, _ := fs.Create("/test.txt")
	f.Close()

	// Now open it
	f, err := fs.Open("/test.txt")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer f.Close()

	if f.Name() != "/test.txt" {
		t.Errorf("expected name /test.txt, got %s", f.Name())
	}
}

func TestFileSystemMkdir(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	err := fs.Mkdir("/testdir", 0755)
	if err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	info, err := fs.Stat("/testdir")
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestFileSystemMkdirAll(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	err := fs.MkdirAll("/a/b/c", 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Verify intermediate directories were created
	for _, path := range []string{"/a", "/a/b", "/a/b/c"} {
		info, err := fs.Stat(path)
		if err != nil {
			t.Errorf("Stat(%s) failed: %v", path, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected %s to be a directory", path)
		}
	}
}

func TestFileSystemRemove(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	fs.Create("/test.txt")

	err := fs.Remove("/test.txt")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	_, err = fs.Stat("/test.txt")
	if err == nil {
		t.Error("file should not exist after removal")
	}
}

func TestFileSystemRename(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	fs.Create("/old.txt")

	err := fs.Rename("/old.txt", "/new.txt")
	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	_, err = fs.Stat("/new.txt")
	if err != nil {
		t.Error("renamed file should exist")
	}

	_, err = fs.Stat("/old.txt")
	if err == nil {
		t.Error("old file should not exist after rename")
	}
}

func TestFileSystemChmod(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	fs.Create("/test.txt")

	err := fs.Chmod("/test.txt", 0644)
	if err != nil {
		t.Fatalf("Chmod failed: %v", err)
	}

	info, _ := fs.Stat("/test.txt")
	if info.Mode() != 0644 {
		t.Errorf("expected mode 0644, got %o", info.Mode())
	}
}

func TestFileSystemChtimes(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	fs.Create("/test.txt")

	newTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	err := fs.Chtimes("/test.txt", newTime, newTime)
	if err != nil {
		t.Fatalf("Chtimes failed: %v", err)
	}

	info, _ := fs.Stat("/test.txt")
	if !info.ModTime().Equal(newTime) {
		t.Errorf("expected modtime %v, got %v", newTime, info.ModTime())
	}
}

func TestFileSystemChown(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	fs.Create("/test.txt")

	err := fs.Chown("/test.txt", 1000, 1000)
	if err != nil {
		t.Fatalf("Chown failed: %v", err)
	}
}

func TestFileSystemSeparator(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	sep := fs.Separator()
	if sep != filepath.Separator {
		t.Errorf("expected separator %c, got %c", filepath.Separator, sep)
	}
}

func TestFileSystemListSeparator(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	sep := fs.ListSeparator()
	if sep != filepath.ListSeparator {
		t.Errorf("expected list separator %c, got %c", filepath.ListSeparator, sep)
	}
}

func TestFileSystemChdir(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	// Create a directory
	fs.Mkdir("/testdir", 0755)

	err := fs.Chdir("/testdir")
	if err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	cwd, err := fs.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}

	if cwd != "/testdir" {
		t.Errorf("expected cwd /testdir, got %s", cwd)
	}
}

func TestFileSystemChdirNonDirectory(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	fs.Create("/file.txt")

	err := fs.Chdir("/file.txt")
	if err == nil {
		t.Error("Chdir on file should fail")
	}
}

func TestFileSystemGetwd(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	cwd, err := fs.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}

	if cwd != "/" {
		t.Errorf("expected initial cwd /, got %s", cwd)
	}
}

func TestFileSystemTempDir(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	tmpDir := fs.TempDir()
	if tmpDir == "" {
		t.Error("TempDir returned empty string")
	}
}

func TestFileSystemTruncate(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	// Create and write to file
	f, _ := fs.Create("/test.txt")
	f.Write([]byte("hello world"))
	f.Close()

	// Truncate it
	err := fs.Truncate("/test.txt", 0)
	if err != nil {
		t.Fatalf("Truncate failed: %v", err)
	}

	info, _ := fs.Stat("/test.txt")
	if info.Size() != 0 {
		t.Errorf("expected size 0 after truncate, got %d", info.Size())
	}
}

func TestFileSystemRelativePaths(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	// Create directory structure
	fs.MkdirAll("/home/user", 0755)
	fs.Chdir("/home/user")

	// Create file with relative path
	f, err := fs.Create("test.txt")
	if err != nil {
		t.Fatalf("Create with relative path failed: %v", err)
	}
	f.Close()

	// Verify it was created in the right place
	_, err = fs.Stat("/home/user/test.txt")
	if err != nil {
		t.Error("file should exist at /home/user/test.txt")
	}
}

// RemoveAll tests are complex and require a more sophisticated mock
// They are covered indirectly by other filesystem implementation tests

func TestReadOnlyFilerInterface(t *testing.T) {
	// This test just verifies the interface compiles
	var _ ReadOnlyFiler = (*mockReadOnlyFiler)(nil)
}

type mockReadOnlyFiler struct{}

func (m *mockReadOnlyFiler) Open(name string) (io.ReadCloser, error) {
	return nil, nil
}

func TestWriteOnlyFilerInterface(t *testing.T) {
	// This test just verifies the interface compiles
	var _ WriteOnlyFiler = (*mockWriteOnlyFiler)(nil)
}

type mockWriteOnlyFiler struct{}

func (m *mockWriteOnlyFiler) Open(name string) (io.WriteCloser, error) {
	return nil, nil
}
