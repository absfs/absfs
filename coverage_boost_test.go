package absfs

import (
	"io"
	"os"
	"testing"
	"time"
)

// Tests to boost coverage for remaining uncovered paths

func TestParseFlags_AllCombinations(t *testing.T) {
	tests := []struct {
		input    string
		expected Flags
		wantErr  bool
	}{
		{"O_RDONLY", Flags(O_RDONLY), false},
		{"O_WRONLY|O_CREATE", Flags(O_WRONLY | O_CREATE), false},
		{"O_RDWR|O_APPEND|O_SYNC", Flags(O_RDWR | O_APPEND | O_SYNC), false},
		{"O_RDONLY|O_RDWR", 0, true}, // Multiple access modes - error
		{"O_WRONLY|O_RDONLY", 0, true}, // Multiple access modes - error
		{"O_INVALID", 0, true},
		{"O_CREATE|O_EXCL|O_TRUNC", Flags(O_CREATE | O_EXCL | O_TRUNC), false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			flags, err := ParseFlags(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for %q", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for %q: %v", tt.input, err)
				}
				if flags != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, flags)
				}
			}
		})
	}
}

func TestFileSystem_OpenWithRelativePath(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	// Create a directory and file
	fs.Mkdir("/testdir", 0755)
	fs.Chdir("/testdir")
	fs.Create("/testdir/file.txt")

	// Test Open with relative path (tests the !filepath.IsAbs branch)
	_, err := fs.Open("file.txt")
	if err != nil {
		t.Errorf("Open with relative path failed: %v", err)
	}
}

func TestFileSystem_CreateWithRelativePath(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	fs.Mkdir("/testdir", 0755)
	fs.Chdir("/testdir")

	// Test Create with relative path
	f, err := fs.Create("newfile.txt")
	if err != nil {
		t.Fatalf("Create with relative path failed: %v", err)
	}
	f.Close()

	// Verify it was created
	_, err = fs.Stat("/testdir/newfile.txt")
	if err != nil {
		t.Error("file should exist")
	}
}

func TestFileSystem_MkdirAll_EdgeCases(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	// Test with root path
	err := fs.MkdirAll("/", 0755)
	if err != nil {
		t.Errorf("MkdirAll(/) should not error: %v", err)
	}

	// Test with relative path
	fs.Chdir("/")
	err = fs.MkdirAll("rel/path/test", 0755)
	if err != nil {
		t.Errorf("MkdirAll with relative path failed: %v", err)
	}
}

func TestFileSystem_TruncateWithoutOptionalInterface(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	// Create file
	f, _ := fs.Create("/test.txt")
	f.Write([]byte("hello world"))
	f.Close()

	// Truncate should work even without optional interface
	err := fs.Truncate("/test.txt", 0)
	if err != nil {
		t.Errorf("Truncate failed: %v", err)
	}
}

func TestFileSystem_RemoveAllWithoutOptionalInterface(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	// Create a simple file
	fs.Create("/test.txt")

	// RemoveAll on a file
	err := fs.RemoveAll("/test.txt")
	if err != nil {
		t.Errorf("RemoveAll on file failed: %v", err)
	}

	// Verify it's removed
	_, err = fs.Stat("/test.txt")
	if err == nil {
		t.Error("file should be removed")
	}
}

func TestFileSystem_AllMethodsWithRelativePaths(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	// Setup
	fs.MkdirAll("/workdir", 0755)
	fs.Chdir("/workdir")

	// Test OpenFile with relative path
	f, err := fs.OpenFile("test.txt", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Errorf("OpenFile with relative path failed: %v", err)
	}
	f.Close()

	// Test Mkdir with relative path
	err = fs.Mkdir("subdir", 0755)
	if err != nil {
		t.Errorf("Mkdir with relative path failed: %v", err)
	}

	// Test Rename with relative paths
	err = fs.Rename("test.txt", "renamed.txt")
	if err != nil {
		t.Errorf("Rename with relative paths failed: %v", err)
	}

	// Test Stat with relative path
	_, err = fs.Stat("renamed.txt")
	if err != nil {
		t.Errorf("Stat with relative path failed: %v", err)
	}

	// Test Chmod with relative path
	err = fs.Chmod("renamed.txt", 0600)
	if err != nil {
		t.Errorf("Chmod with relative path failed: %v", err)
	}

	// Test Chown with relative path
	err = fs.Chown("renamed.txt", 0, 0)
	if err != nil {
		t.Errorf("Chown with relative path failed: %v", err)
	}

	// Test Remove with relative path
	err = fs.Remove("renamed.txt")
	if err != nil {
		t.Errorf("Remove with relative path failed: %v", err)
	}
}

func TestFileAdapter_SeekErrors(t *testing.T) {
	ts := &testSeekable{
		name: "test.txt",
		data: []byte("hello"),
	}

	f := ExtendSeekable(ts)

	// Test ReadAt that requires seek
	buf := make([]byte, 3)
	n, err := f.ReadAt(buf, 2)
	if err != nil && err != io.EOF {
		t.Errorf("ReadAt failed: %v", err)
	}
	if string(buf[:n]) != "llo" {
		t.Errorf("expected 'llo', got '%s'", string(buf[:n]))
	}

	// Test WriteAt that requires seek
	n, err = f.WriteAt([]byte("XX"), 1)
	if err != nil {
		t.Errorf("WriteAt failed: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 bytes written, got %d", n)
	}
}

func TestFileAdapter_TruncateMultipleSizes(t *testing.T) {
	// testSeekable without filetruncater interface uses Write to expand
	ts := &testSeekable{
		name: "test.txt",
		data: []byte{},
	}

	f := ExtendSeekable(ts)

	// Truncate expands by writing zeros (since testSeekable doesn't implement Truncate)
	err := f.Truncate(10)
	if err != nil {
		t.Errorf("Truncate to 10 failed: %v", err)
	}
	// Since it writes zeros, length should be 10
	if len(ts.data) != 10 {
		t.Errorf("expected length 10, got %d", len(ts.data))
	}

	// Write more data
	ts.offset = 10
	for i := 0; i < 5; i++ {
		_, err := f.Write([]byte("x"))
		if err != nil {
			t.Errorf("Write iteration %d failed: %v", i, err)
		}
	}
}

func TestFileSystem_ChdirErrors(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	// Test Chdir to non-existent directory
	err := fs.Chdir("/nonexistent")
	if err == nil {
		t.Error("Chdir to nonexistent directory should fail")
	}

	// Test Chdir to a file (not a directory)
	fs.Create("/file.txt")
	err = fs.Chdir("/file.txt")
	if err == nil {
		t.Error("Chdir to file should fail")
	}
}

// Test edge cases in the mock implementations to ensure they work correctly
func TestMockImplementations(t *testing.T) {
	// Test mockFile interface compliance
	var _ os.FileInfo = (*mockFile)(nil)

	// Test mockFileHandle interface compliance
	mf := &mockFile{name: "test"}
	mfh := &mockFileHandle{file: mf, offset: 0}

	// Test all File interface methods
	var _ File = mfh

	// Test reads and writes with offsets
	mfh.file.content = []byte("hello world")
	mfh.offset = 6

	buf := make([]byte, 5)
	n, err := mfh.Read(buf)
	if err != nil && err != io.EOF {
		t.Errorf("Read failed: %v", err)
	}
	if string(buf[:n]) != "world" {
		t.Errorf("expected 'world', got '%s'", string(buf[:n]))
	}

	// Test write at offset
	mfh.offset = 0
	n, err = mfh.Write([]byte("TEST"))
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}
	if n != 4 {
		t.Errorf("expected 4 bytes written, got %d", n)
	}

	// Test ReadAt and WriteAt
	buf = make([]byte, 4)
	n, err = mfh.ReadAt(buf, 0)
	if err != nil && err != io.EOF {
		t.Errorf("ReadAt failed: %v", err)
	}
	if string(buf[:n]) != "TEST" {
		t.Errorf("expected 'TEST', got '%s'", string(buf[:n]))
	}

	// Test WriteString
	mfh.offset = 0
	n, err = mfh.WriteString("NEWDATA")
	if err != nil {
		t.Errorf("WriteString failed: %v", err)
	}
	if n != 7 {
		t.Errorf("expected 7 bytes written, got %d", n)
	}

	// Test Truncate
	err = mfh.Truncate(3)
	if err != nil {
		t.Errorf("Truncate failed: %v", err)
	}
	if len(mfh.file.content) != 3 {
		t.Errorf("expected length 3, got %d", len(mfh.file.content))
	}

	// Test Seek
	pos, err := mfh.Seek(2, io.SeekStart)
	if err != nil {
		t.Errorf("Seek failed: %v", err)
	}
	if pos != 2 {
		t.Errorf("expected position 2, got %d", pos)
	}
}

func TestExtendFilerCoverage(t *testing.T) {
	// Test that ExtendFiler creates proper wrapper
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	// Verify it implements FileSystem
	var _ FileSystem = fs

	// Test initial working directory
	cwd, err := fs.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	if cwd != "/" {
		t.Errorf("expected initial cwd '/', got '%s'", cwd)
	}
}

func TestFileSystem_RenameRelativePaths(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	fs.MkdirAll("/dir", 0755)
	fs.Chdir("/dir")
	fs.Create("/dir/old.txt")

	// Rename with relative paths
	err := fs.Rename("old.txt", "new.txt")
	if err != nil {
		t.Errorf("Rename with relative paths failed: %v", err)
	}

	_, err = fs.Stat("/dir/new.txt")
	if err != nil {
		t.Error("renamed file should exist")
	}
}

func TestFileSystem_OpenFileRelativePath(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	fs.MkdirAll("/test", 0755)
	fs.Chdir("/test")

	// OpenFile with relative path
	f, err := fs.OpenFile("file.txt", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("OpenFile with relative path failed: %v", err)
	}
	f.Close()

	// Verify
	_, err = fs.Stat("/test/file.txt")
	if err != nil {
		t.Error("file should exist at absolute path")
	}
}

func TestFileSystem_ChtimesRelativePath(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	fs.MkdirAll("/test", 0755)
	fs.Chdir("/test")
	fs.Create("/test/file.txt")

	// Chtimes with relative path
	err := fs.Chtimes("file.txt", testTime, testTime)
	if err != nil {
		t.Errorf("Chtimes with relative path failed: %v", err)
	}
}

// testTime is a test timestamp
var testTime = time.Now()
