package absfs

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestInvalidFile(t *testing.T) {
	f := &InvalidFile{Path: "/test/file.txt"}

	t.Run("Name", func(t *testing.T) {
		if f.Name() != filepath.Clean("/test/file.txt") {
			t.Errorf("expected %s, got %s", filepath.Clean("/test/file.txt"), f.Name())
		}
	})

	t.Run("Read", func(t *testing.T) {
		buf := make([]byte, 10)
		n, err := f.Read(buf)
		if n != 0 {
			t.Errorf("expected 0 bytes read, got %d", n)
		}
		if err == nil {
			t.Error("expected error, got nil")
		}
		pathErr, ok := err.(*os.PathError)
		if !ok {
			t.Errorf("expected *os.PathError, got %T", err)
		} else {
			if pathErr.Op != "read" {
				t.Errorf("expected op 'read', got %s", pathErr.Op)
			}
			if pathErr.Err != syscall.EBADF {
				t.Errorf("expected EBADF, got %v", pathErr.Err)
			}
		}
	})

	t.Run("Write", func(t *testing.T) {
		buf := []byte("test")
		n, err := f.Write(buf)
		if n != 0 {
			t.Errorf("expected 0 bytes written, got %d", n)
		}
		if err == nil {
			t.Error("expected error, got nil")
		}
		pathErr, ok := err.(*os.PathError)
		if !ok {
			t.Errorf("expected *os.PathError, got %T", err)
		} else {
			if pathErr.Op != "write" {
				t.Errorf("expected op 'write', got %s", pathErr.Op)
			}
		}
	})

	t.Run("Close", func(t *testing.T) {
		err := f.Close()
		if err != nil {
			t.Errorf("Close should not return error, got %v", err)
		}
	})

	t.Run("Sync", func(t *testing.T) {
		err := f.Sync()
		if err == nil {
			t.Error("expected error, got nil")
		}
		pathErr, ok := err.(*os.PathError)
		if !ok {
			t.Errorf("expected *os.PathError, got %T", err)
		} else {
			if pathErr.Op != "sync" {
				t.Errorf("expected op 'sync', got %s", pathErr.Op)
			}
		}
	})

	t.Run("Stat", func(t *testing.T) {
		info, err := f.Stat()
		if info != nil {
			t.Error("expected nil FileInfo")
		}
		if err == nil {
			t.Error("expected error, got nil")
		}
		pathErr, ok := err.(*os.PathError)
		if !ok {
			t.Errorf("expected *os.PathError, got %T", err)
		} else {
			if pathErr.Op != "stat" {
				t.Errorf("expected op 'stat', got %s", pathErr.Op)
			}
		}
	})

	t.Run("Readdir", func(t *testing.T) {
		infos, err := f.Readdir(10)
		if infos != nil {
			t.Error("expected nil slice")
		}
		if err == nil {
			t.Error("expected error, got nil")
		}
		pathErr, ok := err.(*os.PathError)
		if !ok {
			t.Errorf("expected *os.PathError, got %T", err)
		} else {
			if pathErr.Op != "readdir" {
				t.Errorf("expected op 'readdir', got %s", pathErr.Op)
			}
		}
	})

	t.Run("Seek", func(t *testing.T) {
		pos, err := f.Seek(0, 0)
		if pos != 0 {
			t.Errorf("expected position 0, got %d", pos)
		}
		if err == nil {
			t.Error("expected error, got nil")
		}
		pathErr, ok := err.(*os.PathError)
		if !ok {
			t.Errorf("expected *os.PathError, got %T", err)
		} else {
			if pathErr.Op != "seek" {
				t.Errorf("expected op 'seek', got %s", pathErr.Op)
			}
		}
	})

	t.Run("ReadAt", func(t *testing.T) {
		buf := make([]byte, 10)
		n, err := f.ReadAt(buf, 0)
		if n != 0 {
			t.Errorf("expected 0 bytes read, got %d", n)
		}
		if err == nil {
			t.Error("expected error, got nil")
		}
		pathErr, ok := err.(*os.PathError)
		if !ok {
			t.Errorf("expected *os.PathError, got %T", err)
		} else {
			if pathErr.Op != "read" {
				t.Errorf("expected op 'read', got %s", pathErr.Op)
			}
		}
	})

	t.Run("WriteAt", func(t *testing.T) {
		buf := []byte("test")
		n, err := f.WriteAt(buf, 0)
		if n != 0 {
			t.Errorf("expected 0 bytes written, got %d", n)
		}
		if err == nil {
			t.Error("expected error, got nil")
		}
		pathErr, ok := err.(*os.PathError)
		if !ok {
			t.Errorf("expected *os.PathError, got %T", err)
		} else {
			if pathErr.Op != "write" {
				t.Errorf("expected op 'write', got %s", pathErr.Op)
			}
		}
	})

	t.Run("WriteString", func(t *testing.T) {
		n, err := f.WriteString("test")
		if n != 0 {
			t.Errorf("expected 0 bytes written, got %d", n)
		}
		if err == nil {
			t.Error("expected error, got nil")
		}
		pathErr, ok := err.(*os.PathError)
		if !ok {
			t.Errorf("expected *os.PathError, got %T", err)
		} else {
			if pathErr.Op != "write" {
				t.Errorf("expected op 'write', got %s", pathErr.Op)
			}
		}
	})

	t.Run("Truncate", func(t *testing.T) {
		err := f.Truncate(0)
		if err == nil {
			t.Error("expected error, got nil")
		}
		pathErr, ok := err.(*os.PathError)
		if !ok {
			t.Errorf("expected *os.PathError, got %T", err)
		} else {
			if pathErr.Op != "truncate" {
				t.Errorf("expected op 'truncate', got %s", pathErr.Op)
			}
		}
	})

	t.Run("Readdirnames", func(t *testing.T) {
		names, err := f.Readdirnames(10)
		if names != nil {
			t.Error("expected nil slice")
		}
		if err == nil {
			t.Error("expected error, got nil")
		}
		pathErr, ok := err.(*os.PathError)
		if !ok {
			t.Errorf("expected *os.PathError, got %T", err)
		} else {
			if pathErr.Op != "readdirnames" {
				t.Errorf("expected op 'readdirnames', got %s", pathErr.Op)
			}
		}
	})
}
