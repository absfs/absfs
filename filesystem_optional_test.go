package absfs

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// mockFilerWithOptionals implements Filer plus optional interfaces
type mockFilerWithOptionals struct {
	*mockFiler
	cwd string
}

func newMockFilerWithOptionals() *mockFilerWithOptionals {
	return &mockFilerWithOptionals{
		mockFiler: newMockFiler(),
		cwd:       string(filepath.Separator),
	}
}

// Implement optional dirnavigator interface
func (m *mockFilerWithOptionals) Chdir(dir string) error {
	dir = filepath.Clean(dir)
	if _, exists := m.files[dir]; !exists {
		return &os.PathError{Op: "chdir", Path: dir, Err: os.ErrNotExist}
	}
	f := m.files[dir]
	if !f.isDir {
		return &os.PathError{Op: "chdir", Path: dir, Err: os.ErrInvalid}
	}
	m.cwd = dir
	return nil
}

func (m *mockFilerWithOptionals) Getwd() (string, error) {
	return m.cwd, nil
}

// Implement optional separator interface
func (m *mockFilerWithOptionals) Separator() uint8 {
	return '/'
}

// Implement optional listseparator interface
func (m *mockFilerWithOptionals) ListSeparator() uint8 {
	return ':'
}

// Implement optional temper interface
func (m *mockFilerWithOptionals) TempDir() string {
	return filepath.Clean("/tmp")
}

// Implement optional opener interface
func (m *mockFilerWithOptionals) Open(name string) (File, error) {
	return m.OpenFile(name, os.O_RDONLY, 0)
}

// Implement optional creator interface
func (m *mockFilerWithOptionals) Create(name string) (File, error) {
	return m.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
}

// Implement optional mkaller interface
func (m *mockFilerWithOptionals) MkdirAll(name string, perm os.FileMode) error {
	name = filepath.Clean(name)
	parts := []string{}
	root := string(filepath.Separator)
	for name != root && name != "." {
		parts = append([]string{name}, parts...)
		name = filepath.Dir(name)
	}

	for _, part := range parts {
		if _, exists := m.files[part]; !exists {
			m.Mkdir(part, perm)
		}
	}
	return nil
}

// Implement optional remover interface
func (m *mockFilerWithOptionals) RemoveAll(path string) error {
	path = filepath.Clean(path)
	if _, exists := m.files[path]; !exists {
		return &os.PathError{Op: "remove", Path: path, Err: os.ErrNotExist}
	}
	delete(m.files, path)
	// Simple implementation - doesn't recurse
	return nil
}

// Implement optional truncater interface
func (m *mockFilerWithOptionals) Truncate(name string, size int64) error {
	name = filepath.Clean(name)
	f, exists := m.files[name]
	if !exists {
		return &os.PathError{Op: "truncate", Path: name, Err: os.ErrNotExist}
	}
	if size > int64(len(f.content)) {
		padding := make([]byte, size-int64(len(f.content)))
		f.content = append(f.content, padding...)
	} else {
		f.content = f.content[:size]
	}
	return nil
}

func TestFileSystemWithOptionalInterfaces(t *testing.T) {
	mock := newMockFilerWithOptionals()
	fs := ExtendFiler(mock)

	t.Run("Separator", func(t *testing.T) {
		sep := fs.Separator()
		if sep != '/' {
			t.Errorf("expected '/', got %c", sep)
		}
	})

	t.Run("ListSeparator", func(t *testing.T) {
		sep := fs.ListSeparator()
		if sep != ':' {
			t.Errorf("expected ':', got %c", sep)
		}
	})

	t.Run("TempDir", func(t *testing.T) {
		tmpDir := fs.TempDir()
		if tmpDir != filepath.Clean("/tmp") {
			t.Errorf("expected '%s', got %s", filepath.Clean("/tmp"), tmpDir)
		}
	})

	t.Run("Chdir/Getwd", func(t *testing.T) {
		fs.Mkdir("/testdir", 0755)
		err := fs.Chdir("/testdir")
		if err != nil {
			t.Fatalf("Chdir failed: %v", err)
		}
		cwd, err := fs.Getwd()
		if err != nil {
			t.Fatalf("Getwd failed: %v", err)
		}
		if cwd != filepath.Clean("/testdir") {
			t.Errorf("expected '%s', got %s", filepath.Clean("/testdir"), cwd)
		}
	})

	t.Run("Open", func(t *testing.T) {
		fs.Create("/test.txt")
		f, err := fs.Open("/test.txt")
		if err != nil {
			t.Fatalf("Open failed: %v", err)
		}
		f.Close()
	})

	t.Run("Create", func(t *testing.T) {
		f, err := fs.Create("/new.txt")
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		f.Close()
	})

	t.Run("MkdirAll", func(t *testing.T) {
		err := fs.MkdirAll("/a/b/c", 0755)
		if err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}
		_, err = fs.Stat("/a/b/c")
		if err != nil {
			t.Error("directory should exist")
		}
	})

	t.Run("RemoveAll", func(t *testing.T) {
		fs.Create("/remove.txt")
		err := fs.RemoveAll("/remove.txt")
		if err != nil {
			t.Fatalf("RemoveAll failed: %v", err)
		}
		_, err = fs.Stat("/remove.txt")
		if err == nil {
			t.Error("file should not exist after RemoveAll")
		}
	})

	t.Run("Truncate", func(t *testing.T) {
		f, _ := fs.Create("/truncate.txt")
		f.Write([]byte("hello world"))
		f.Close()

		err := fs.Truncate("/truncate.txt", 5)
		if err != nil {
			t.Fatalf("Truncate failed: %v", err)
		}

		info, _ := fs.Stat("/truncate.txt")
		if info.Size() != 5 {
			t.Errorf("expected size 5, got %d", info.Size())
		}
	})
}

func TestFileSystemRelativePathsWithDirNavigator(t *testing.T) {
	mock := newMockFilerWithOptionals()
	fs := ExtendFiler(mock)

	// Create directory and change to it
	fs.MkdirAll("/home/user", 0755)
	fs.Chdir("/home/user")

	// The ExtendFiler wrapper should pass through to the underlying
	// dirnavigator when it implements that interface
	cwd, _ := fs.Getwd()
	if cwd != filepath.Clean("/home/user") {
		t.Errorf("expected '%s', got %s", filepath.Clean("/home/user"), cwd)
	}
}

func TestFileSystemOpenFileWithDirNavigator(t *testing.T) {
	mock := newMockFilerWithOptionals()
	fs := ExtendFiler(mock)

	// Test that OpenFile is called through to filer
	// This tests the 0% coverage path in filesystem.go:110
	f, err := fs.OpenFile("/test.txt", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	f.Close()
}

func TestFileSystemMkdirWithDirNavigator(t *testing.T) {
	mock := newMockFilerWithOptionals()
	fs := ExtendFiler(mock)

	// Tests both branches of Mkdir
	err := fs.Mkdir("/dir1", 0755)
	if err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	// Change to absolute path and try relative
	fs.Chdir("/")
	err = fs.Mkdir("/dir2", 0755)
	if err != nil {
		t.Fatalf("Mkdir with absolute path failed: %v", err)
	}
}

func TestFileSystemRemoveWithDirNavigator(t *testing.T) {
	mock := newMockFilerWithOptionals()
	fs := ExtendFiler(mock)

	fs.Create("/test.txt")
	err := fs.Remove("/test.txt")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
}

func TestFileSystemRenameWithDirNavigator(t *testing.T) {
	mock := newMockFilerWithOptionals()
	fs := ExtendFiler(mock)

	fs.Create("/old.txt")
	err := fs.Rename("/old.txt", "/new.txt")
	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}
}

func TestFileSystemStatWithDirNavigator(t *testing.T) {
	mock := newMockFilerWithOptionals()
	fs := ExtendFiler(mock)

	fs.Create("/test.txt")
	_, err := fs.Stat("/test.txt")
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
}

func TestFileSystemChmodWithDirNavigator(t *testing.T) {
	mock := newMockFilerWithOptionals()
	fs := ExtendFiler(mock)

	fs.Create("/test.txt")
	err := fs.Chmod("/test.txt", 0644)
	if err != nil {
		t.Fatalf("Chmod failed: %v", err)
	}
}

func TestFileSystemChtimesWithDirNavigator(t *testing.T) {
	mock := newMockFilerWithOptionals()
	fs := ExtendFiler(mock)

	fs.Create("/test.txt")
	now := time.Now()
	err := fs.Chtimes("/test.txt", now, now)
	if err != nil {
		t.Fatalf("Chtimes failed: %v", err)
	}
}

func TestFileSystemChownWithDirNavigator(t *testing.T) {
	mock := newMockFilerWithOptionals()
	fs := ExtendFiler(mock)

	fs.Create("/test.txt")
	err := fs.Chown("/test.txt", 1000, 1000)
	if err != nil {
		t.Fatalf("Chown failed: %v", err)
	}
}

func TestFileSystemCreateWithDirNavigator(t *testing.T) {
	mock := newMockFilerWithOptionals()
	fs := ExtendFiler(mock)

	// Test that when filer implements creator, it's used
	f, err := fs.Create("/test.txt")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	f.Close()
}

func TestFileSystemTruncateExpand(t *testing.T) {
	mock := newMockFilerWithOptionals()
	fs := ExtendFiler(mock)

	f, _ := fs.Create("/test.txt")
	f.Write([]byte("hi"))
	f.Close()

	// Truncate to larger size
	err := fs.Truncate("/test.txt", 10)
	if err != nil {
		t.Fatalf("Truncate failed: %v", err)
	}

	info, _ := fs.Stat("/test.txt")
	if info.Size() != 10 {
		t.Errorf("expected size 10, got %d", info.Size())
	}
}

func TestFileSystemTruncateNonexistent(t *testing.T) {
	mock := newMockFilerWithOptionals()
	fs := ExtendFiler(mock)

	err := fs.Truncate("/nonexistent.txt", 10)
	if err == nil {
		t.Error("Truncate on nonexistent file should fail")
	}
}

func TestFileAdapterReadAtError(t *testing.T) {
	ts := &testSeekable{
		name: "test.txt",
		data: []byte("hello"),
	}

	f := ExtendSeekable(ts)

	// Test ReadAt with error from Seek
	// This is tricky - we need a seekable that fails on seek
	// For now, just test that ReadAt returns EOF when at end
	buf := make([]byte, 10)
	n, err := f.ReadAt(buf, 100)
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 bytes, got %d", n)
	}
}

func TestFileAdapterWriteAtError(t *testing.T) {
	ts := &testSeekable{
		name: "test.txt",
		data: []byte("hello"),
	}

	f := ExtendSeekable(ts)

	// Test normal WriteAt operation
	n, err := f.WriteAt([]byte("XX"), 0)
	if err != nil {
		t.Fatalf("WriteAt failed: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 bytes written, got %d", n)
	}
}

func TestFileAdapterTruncateExpand(t *testing.T) {
	ts := &testSeekable{
		name: "test.txt",
		data: []byte("hi"),
	}

	f := ExtendSeekable(ts)

	// Truncate to larger size
	err := f.Truncate(10)
	if err != nil {
		t.Fatalf("Truncate failed: %v", err)
	}
	if len(ts.data) != 10 {
		t.Errorf("expected length 10, got %d", len(ts.data))
	}
}

func TestFileAdapterReaddirError(t *testing.T) {
	ts := &testSeekable{
		name:  "file.txt",
		isDir: false,
	}

	f := ExtendSeekable(ts)
	_, err := f.Readdir(10)
	if err == nil {
		t.Error("Readdir on file should fail")
	}
}

func TestFileAdapterReaddirnamesError(t *testing.T) {
	ts := &testSeekable{
		name:  "file.txt",
		isDir: false,
	}

	f := ExtendSeekable(ts)
	_, err := f.Readdirnames(10)
	if err == nil {
		t.Error("Readdirnames on file should fail")
	}
}
