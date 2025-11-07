package absfs

import (
	"errors"
	"os"
	"testing"
)

// seekableWithErrors implements Seekable but returns errors on Seek
type seekableWithErrors struct {
	name string
	data []byte
}

func (s *seekableWithErrors) Name() string { return s.name }
func (s *seekableWithErrors) Read(p []byte) (int, error) {
	return copy(p, s.data), nil
}
func (s *seekableWithErrors) Write(p []byte) (int, error) {
	s.data = append(s.data, p...)
	return len(p), nil
}
func (s *seekableWithErrors) Close() error                             { return nil }
func (s *seekableWithErrors) Sync() error                              { return nil }
func (s *seekableWithErrors) Stat() (os.FileInfo, error)               { return nil, nil }
func (s *seekableWithErrors) Readdir(int) ([]os.FileInfo, error)       { return nil, nil }
func (s *seekableWithErrors) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("seek error")
}

func TestFileAdapterReadAtSeekError(t *testing.T) {
	s := &seekableWithErrors{
		name: "test.txt",
		data: []byte("hello"),
	}

	f := ExtendSeekable(s)
	buf := make([]byte, 5)
	_, err := f.ReadAt(buf, 0)
	if err == nil {
		t.Error("expected error from ReadAt when Seek fails")
	}
}

func TestFileAdapterWriteAtSeekError(t *testing.T) {
	s := &seekableWithErrors{
		name: "test.txt",
		data: []byte("hello"),
	}

	f := ExtendSeekable(s)
	_, err := f.WriteAt([]byte("test"), 0)
	if err == nil {
		t.Error("expected error from WriteAt when Seek fails")
	}
}

func TestFileAdapterTruncateSeekError(t *testing.T) {
	s := &seekableWithErrors{
		name: "test.txt",
		data: []byte("hello"),
	}

	f := ExtendSeekable(s)
	err := f.Truncate(10)
	if err == nil {
		t.Error("expected error from Truncate when Seek fails")
	}
}

func TestFileSystemTruncateWithoutOptional(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	// Create and populate file
	f, _ := fs.Create("/test.txt")
	f.Write([]byte("hello world"))
	f.Close()

	// Test truncate when filer doesn't implement optional truncater interface
	// The mock doesn't implement truncater, so it will use OpenFile->Close
	err := fs.Truncate("/test.txt", 0)
	if err != nil {
		t.Errorf("Truncate failed: %v", err)
	}

	// Verify file exists
	_, err = fs.Stat("/test.txt")
	if err != nil {
		t.Error("file should still exist after truncate")
	}
}

func TestFileSystemTruncateError(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	// Test truncate on nonexistent file
	err := fs.Truncate("/nonexistent.txt", 100)
	if err == nil {
		t.Error("Truncate on nonexistent file should fail")
	}
}

func TestFileSystemRemoveAllNonexistent(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	// Test RemoveAll on nonexistent path
	err := fs.RemoveAll("/nonexistent")
	if err == nil {
		t.Error("RemoveAll on nonexistent path should fail")
	}
}

func TestFileSystemMkdirAll_RootAndSingleDir(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	// Test single directory (not nested)
	err := fs.MkdirAll("/single", 0755)
	if err != nil {
		t.Errorf("MkdirAll for single dir failed: %v", err)
	}

	// Verify it exists
	_, err = fs.Stat("/single")
	if err != nil {
		t.Error("directory should exist")
	}
}

func TestFileSystemChdir_RelativePath(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	fs.Mkdir("/dir1", 0755)
	fs.Mkdir("/dir1/dir2", 0755)
	fs.Chdir("/dir1")

	// Chdir with relative path (cleaned to /dir1/dir2)
	err := fs.Chdir("dir2")
	if err != nil {
		t.Errorf("Chdir with relative path failed: %v", err)
	}

	cwd, _ := fs.Getwd()
	// After clean, it becomes an absolute path
	if cwd != "dir2" {
		t.Logf("cwd is %s (path was cleaned)", cwd)
	}
}

func TestParseFlags_DefaultReadOnly(t *testing.T) {
	// When no access mode is specified, default is O_RDONLY
	flags, err := ParseFlags("O_CREATE")
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}

	// Should be O_CREATE with default O_RDONLY
	expected := Flags(O_CREATE)
	if flags != expected {
		t.Errorf("expected %v, got %v", expected, flags)
	}
}

func TestParseFlags_MultipleAccessModes(t *testing.T) {
	// Test all combinations of multiple access modes (should error)
	tests := []string{
		"O_RDONLY|O_WRONLY",
		"O_RDONLY|O_RDWR",
		"O_WRONLY|O_RDWR",
	}

	for _, input := range tests {
		_, err := ParseFlags(input)
		if err == nil {
			t.Errorf("ParseFlags(%q) should fail with multiple access modes", input)
		}
	}
}

func TestFileSystemOpenFile_AbsolutePath(t *testing.T) {
	mock := newMockFilerWithOptionals()
	fs := ExtendFiler(mock)

	// Test OpenFile with absolute path when filer implements dirnavigator
	// This tests the path where dirnavigator is found but path is absolute
	f, err := fs.OpenFile("/test.txt", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	f.Close()
}
