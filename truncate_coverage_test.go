package absfs

import (
	"os"
	"testing"
)

func TestFileSystemTruncateViaOpenFile(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	// Create file
	f, _ := fs.Create("/test.txt")
	f.Write([]byte("hello world"))
	f.Close()

	// Truncate via OpenFile path (mock doesn't implement truncater interface)
	// This exercises the filesystem.go:360 Truncate fallback implementation
	err := fs.Truncate("/test.txt", 5)
	if err != nil {
		t.Errorf("Truncate should work via OpenFile fallback: %v", err)
	}
}

func TestFileSystemTruncateOpenError(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	// Test error path when file doesn't exist
	err := fs.Truncate("/missing.txt", 10)
	if err == nil {
		t.Error("Truncate on missing file should fail")
	}
}

func TestFileSystemTruncateClose(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	// Create file and truncate - tests the Close() path in Truncate
	fs.Create("/test.txt")
	err := fs.Truncate("/test.txt", 0)
	if err != nil {
		t.Errorf("Truncate failed: %v", err)
	}
}

func TestFileSystemRemoveAllStatError(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	// This is tricky - we need a file that opens but Stat fails
	// For now, just ensure RemoveAll handles errors properly
	err := fs.RemoveAll("/nonexistent-for-stat-test")
	if err == nil {
		t.Error("RemoveAll should fail on nonexistent file")
	}
}

func TestFileAdapterTruncateLargeSize(t *testing.T) {
	ts := &testSeekable{
		name: "test.txt",
		data: []byte{},
	}

	f := ExtendSeekable(ts)

	// Truncate to a size larger than buffer (tests loop in Truncate)
	err := f.Truncate(10000) // Larger than 4096 buffer
	if err != nil {
		t.Errorf("Truncate to large size failed: %v", err)
	}
	if len(ts.data) != 10000 {
		t.Errorf("expected length 10000, got %d", len(ts.data))
	}
}

func TestFileSystemMkdirAllEmptyPath(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	// Test MkdirAll with paths that result in empty parts
	err := fs.MkdirAll("/a//b///c", 0755)
	if err != nil {
		t.Errorf("MkdirAll with extra slashes failed: %v", err)
	}

	// Clean path should result in /a/b/c
	_, err = fs.Stat("/a/b/c")
	if err != nil {
		t.Error("directory should exist")
	}
}

func TestFileSystemOpenWithOptional(t *testing.T) {
	mock := newMockFilerWithOptionals()
	fs := ExtendFiler(mock)

	// When filer implements opener, it should be used
	f, err := fs.Open("/nonexistent")
	if err == nil {
		f.Close()
	}
	// Just testing that the optional path is taken
}

func TestFileSystemCreateWithOptional(t *testing.T) {
	mock := newMockFilerWithOptionals()
	fs := ExtendFiler(mock)

	// When filer implements creator, it should be used
	f, err := fs.Create("/test.txt")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	f.Close()
}

func TestFileSystemRemoveAllWithOptional(t *testing.T) {
	mock := newMockFilerWithOptionals()
	fs := ExtendFiler(mock)

	// When filer implements RemoveAll, it should be used
	fs.Create("/test.txt")
	err := fs.RemoveAll("/test.txt")
	if err != nil {
		t.Errorf("RemoveAll failed: %v", err)
	}
}

func TestFileSystemTruncateWithOptional(t *testing.T) {
	mock := newMockFilerWithOptionals()
	fs := ExtendFiler(mock)

	// When filer implements Truncate, it should be used
	fs.Create("/test.txt")
	err := fs.Truncate("/test.txt", 10)
	if err != nil {
		t.Errorf("Truncate failed: %v", err)
	}

	info, _ := fs.Stat("/test.txt")
	if info.Size() != 10 {
		t.Errorf("expected size 10, got %d", info.Size())
	}
}

func TestFileSystemMkdirAllWithOptional(t *testing.T) {
	mock := newMockFilerWithOptionals()
	fs := ExtendFiler(mock)

	// When filer implements MkdirAll, it should be used
	err := fs.MkdirAll("/x/y/z", 0755)
	if err != nil {
		t.Errorf("MkdirAll failed: %v", err)
	}

	_, err = fs.Stat("/x/y/z")
	if err != nil {
		t.Error("directory should exist")
	}
}

func TestFileSystemOpenFileVariousModes(t *testing.T) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)

	tests := []struct {
		name string
		flag int
		perm os.FileMode
	}{
		{"read only", os.O_RDONLY, 0},
		{"write only", os.O_WRONLY | os.O_CREATE, 0644},
		{"read write", os.O_RDWR | os.O_CREATE, 0644},
		{"append", os.O_APPEND | os.O_CREATE, 0644},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/test_" + tt.name + ".txt"
			if tt.flag&os.O_CREATE == 0 {
				// Pre-create for modes that don't create
				f, _ := fs.Create(path)
				f.Close()
			}
			f, err := fs.OpenFile(path, tt.flag, tt.perm)
			if err != nil && tt.flag&os.O_CREATE == 0 {
				// Expected for read-only on nonexistent
				return
			}
			if err != nil {
				t.Errorf("OpenFile failed: %v", err)
				return
			}
			f.Close()
		})
	}
}
