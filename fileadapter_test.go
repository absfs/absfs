package absfs

import (
	"io"
	"os"
	"testing"
)

// testSeekable is a simple Seekable implementation for testing
type testSeekable struct {
	name    string
	data    []byte
	offset  int64
	isDir   bool
	entries []os.FileInfo
}

func (t *testSeekable) Name() string { return t.name }

func (t *testSeekable) Read(p []byte) (int, error) {
	if t.offset >= int64(len(t.data)) {
		return 0, io.EOF
	}
	n := copy(p, t.data[t.offset:])
	t.offset += int64(n)
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

func (t *testSeekable) Write(p []byte) (int, error) {
	if t.offset > int64(len(t.data)) {
		padding := make([]byte, t.offset-int64(len(t.data)))
		t.data = append(t.data, padding...)
	}
	if t.offset == int64(len(t.data)) {
		t.data = append(t.data, p...)
	} else {
		end := t.offset + int64(len(p))
		if end > int64(len(t.data)) {
			t.data = append(t.data[:t.offset], p...)
		} else {
			copy(t.data[t.offset:], p)
		}
	}
	t.offset += int64(len(p))
	return len(p), nil
}

func (t *testSeekable) Close() error { return nil }
func (t *testSeekable) Sync() error  { return nil }

func (t *testSeekable) Stat() (os.FileInfo, error) {
	return &mockFile{
		name:    t.name,
		mode:    0644,
		content: t.data,
	}, nil
}

func (t *testSeekable) Readdir(n int) ([]os.FileInfo, error) {
	if !t.isDir {
		return nil, &os.PathError{Op: "readdir", Path: t.name, Err: os.ErrInvalid}
	}
	return t.entries, nil
}

func (t *testSeekable) Seek(offset int64, whence int) (int64, error) {
	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = t.offset + offset
	case io.SeekEnd:
		newOffset = int64(len(t.data)) + offset
	default:
		return 0, &os.PathError{Op: "seek", Path: t.name, Err: os.ErrInvalid}
	}
	if newOffset < 0 {
		return 0, &os.PathError{Op: "seek", Path: t.name, Err: os.ErrInvalid}
	}
	t.offset = newOffset
	return newOffset, nil
}

func TestExtendSeekable(t *testing.T) {
	ts := &testSeekable{
		name: "test.txt",
		data: []byte("hello world"),
	}

	f := ExtendSeekable(ts)
	if f == nil {
		t.Fatal("ExtendSeekable returned nil")
	}

	// Should be able to use as File
	var _ File = f
}

func TestFileAdapterRead(t *testing.T) {
	ts := &testSeekable{
		name: "test.txt",
		data: []byte("hello world"),
	}

	f := ExtendSeekable(ts)
	buf := make([]byte, 5)
	n, err := f.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5 bytes read, got %d", n)
	}
	if string(buf) != "hello" {
		t.Errorf("expected 'hello', got %s", string(buf))
	}
}

func TestFileAdapterWrite(t *testing.T) {
	ts := &testSeekable{
		name: "test.txt",
		data: []byte{},
	}

	f := ExtendSeekable(ts)
	n, err := f.Write([]byte("test"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != 4 {
		t.Errorf("expected 4 bytes written, got %d", n)
	}
	if string(ts.data) != "test" {
		t.Errorf("expected 'test', got %s", string(ts.data))
	}
}

func TestFileAdapterSeek(t *testing.T) {
	ts := &testSeekable{
		name: "test.txt",
		data: []byte("hello world"),
	}

	f := ExtendSeekable(ts)

	// Seek to position 6
	pos, err := f.Seek(6, io.SeekStart)
	if err != nil {
		t.Fatalf("Seek failed: %v", err)
	}
	if pos != 6 {
		t.Errorf("expected position 6, got %d", pos)
	}

	// Read from new position
	buf := make([]byte, 5)
	n, _ := f.Read(buf)
	if string(buf[:n]) != "world" {
		t.Errorf("expected 'world', got %s", string(buf[:n]))
	}
}

func TestFileAdapterReadAt(t *testing.T) {
	ts := &testSeekable{
		name: "test.txt",
		data: []byte("hello world"),
	}

	f := ExtendSeekable(ts)
	buf := make([]byte, 5)
	n, err := f.ReadAt(buf, 6)
	if err != nil && err != io.EOF {
		t.Fatalf("ReadAt failed: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5 bytes read, got %d", n)
	}
	if string(buf) != "world" {
		t.Errorf("expected 'world', got %s", string(buf))
	}
}

func TestFileAdapterWriteAt(t *testing.T) {
	ts := &testSeekable{
		name: "test.txt",
		data: []byte("hello world"),
	}

	f := ExtendSeekable(ts)
	n, err := f.WriteAt([]byte("TEST"), 0)
	if err != nil {
		t.Fatalf("WriteAt failed: %v", err)
	}
	if n != 4 {
		t.Errorf("expected 4 bytes written, got %d", n)
	}
	if string(ts.data) != "TESTo world" {
		t.Errorf("expected 'TESTo world', got %s", string(ts.data))
	}
}

func TestFileAdapterWriteString(t *testing.T) {
	ts := &testSeekable{
		name: "test.txt",
		data: []byte{},
	}

	f := ExtendSeekable(ts)
	n, err := f.WriteString("hello")
	if err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5 bytes written, got %d", n)
	}
	if string(ts.data) != "hello" {
		t.Errorf("expected 'hello', got %s", string(ts.data))
	}
}

func TestFileAdapterTruncate(t *testing.T) {
	ts := &testSeekable{
		name: "test.txt",
		data: []byte{},
	}

	f := ExtendSeekable(ts)

	// Truncate to 10 bytes (should write zeros)
	err := f.Truncate(10)
	if err != nil {
		t.Fatalf("Truncate failed: %v", err)
	}
	if len(ts.data) != 10 {
		t.Errorf("expected length 10, got %d", len(ts.data))
	}
}

func TestFileAdapterReaddirnames(t *testing.T) {
	entry1 := &mockFile{name: "file1.txt"}
	entry2 := &mockFile{name: "file2.txt"}

	ts := &testSeekable{
		name:    "testdir",
		isDir:   true,
		entries: []os.FileInfo{entry1, entry2},
	}

	f := ExtendSeekable(ts)
	names, err := f.Readdirnames(10)
	if err != nil {
		t.Fatalf("Readdirnames failed: %v", err)
	}
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}
	if names[0] != "file1.txt" || names[1] != "file2.txt" {
		t.Errorf("unexpected names: %v", names)
	}
}

func TestFileAdapterName(t *testing.T) {
	ts := &testSeekable{
		name: "test.txt",
	}

	f := ExtendSeekable(ts)
	if f.Name() != "test.txt" {
		t.Errorf("expected 'test.txt', got %s", f.Name())
	}
}

func TestFileAdapterStat(t *testing.T) {
	ts := &testSeekable{
		name: "test.txt",
		data: []byte("hello"),
	}

	f := ExtendSeekable(ts)
	info, err := f.Stat()
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if info.Size() != 5 {
		t.Errorf("expected size 5, got %d", info.Size())
	}
}

func TestFileAdapterReaddir(t *testing.T) {
	entry1 := &mockFile{name: "file1.txt"}
	ts := &testSeekable{
		name:    "testdir",
		isDir:   true,
		entries: []os.FileInfo{entry1},
	}

	f := ExtendSeekable(ts)
	entries, err := f.Readdir(10)
	if err != nil {
		t.Fatalf("Readdir failed: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestFileAdapterClose(t *testing.T) {
	ts := &testSeekable{
		name: "test.txt",
	}

	f := ExtendSeekable(ts)
	err := f.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestFileAdapterSync(t *testing.T) {
	ts := &testSeekable{
		name: "test.txt",
	}

	f := ExtendSeekable(ts)
	err := f.Sync()
	if err != nil {
		t.Errorf("Sync failed: %v", err)
	}
}

// Test that ExtendSeekable returns the same File if already a File
func TestExtendSeekableAlreadyFile(t *testing.T) {
	ts := &testSeekable{
		name: "test.txt",
	}

	f1 := ExtendSeekable(ts)
	f2 := ExtendSeekable(f1)

	// Should return the same object if already a File
	if f1 != f2 {
		t.Error("ExtendSeekable should return the same File if already a File")
	}
}

// Test with a type that implements ReadAt and WriteAt
type testSeekableWithAt struct {
	testSeekable
}

func (t *testSeekableWithAt) ReadAt(b []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, &os.PathError{Op: "read", Path: t.name, Err: os.ErrInvalid}
	}
	if off >= int64(len(t.data)) {
		return 0, io.EOF
	}
	n = copy(b, t.data[off:])
	if n < len(b) {
		err = io.EOF
	}
	return n, err
}

func (t *testSeekableWithAt) WriteAt(b []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, &os.PathError{Op: "write", Path: t.name, Err: os.ErrInvalid}
	}
	if off > int64(len(t.data)) {
		padding := make([]byte, off-int64(len(t.data)))
		t.data = append(t.data, padding...)
	}
	if off+int64(len(b)) > int64(len(t.data)) {
		t.data = append(t.data[:off], b...)
	} else {
		copy(t.data[off:], b)
	}
	return len(b), nil
}

func TestFileAdapterWithNativeAt(t *testing.T) {
	ts := &testSeekableWithAt{
		testSeekable: testSeekable{
			name: "test.txt",
			data: []byte("hello world"),
		},
	}

	f := ExtendSeekable(ts)

	// Test ReadAt uses native implementation
	buf := make([]byte, 5)
	n, _ := f.ReadAt(buf, 6)
	if string(buf[:n]) != "world" {
		t.Errorf("expected 'world', got %s", string(buf[:n]))
	}

	// Test WriteAt uses native implementation
	f.WriteAt([]byte("TEST"), 0)
	if string(ts.data) != "TESTo world" {
		t.Errorf("expected 'TESTo world', got %s", string(ts.data))
	}
}

// Test with a type that implements Truncate
type testSeekableWithTruncate struct {
	testSeekable
	truncated bool
}

func (t *testSeekableWithTruncate) Truncate(size int64) error {
	t.truncated = true
	if size < 0 {
		return &os.PathError{Op: "truncate", Path: t.name, Err: os.ErrInvalid}
	}
	if size > int64(len(t.data)) {
		padding := make([]byte, size-int64(len(t.data)))
		t.data = append(t.data, padding...)
	} else {
		t.data = t.data[:size]
	}
	return nil
}

func TestFileAdapterWithNativeTruncate(t *testing.T) {
	ts := &testSeekableWithTruncate{
		testSeekable: testSeekable{
			name: "test.txt",
			data: []byte("hello world"),
		},
	}

	f := ExtendSeekable(ts)
	f.Truncate(5)

	if !ts.truncated {
		t.Error("native Truncate should have been called")
	}
	if string(ts.data) != "hello" {
		t.Errorf("expected 'hello', got %s", string(ts.data))
	}
}

// Test with a type that implements Readdirnames
type testSeekableWithReaddirnames struct {
	testSeekable
}

func (t *testSeekableWithReaddirnames) Readdirnames(n int) ([]string, error) {
	return []string{"custom1.txt", "custom2.txt"}, nil
}

func TestFileAdapterWithNativeReaddirnames(t *testing.T) {
	ts := &testSeekableWithReaddirnames{
		testSeekable: testSeekable{
			name:  "testdir",
			isDir: true,
			entries: []os.FileInfo{
				&mockFile{name: "ignored.txt"},
			},
		},
	}

	f := ExtendSeekable(ts)
	names, err := f.Readdirnames(10)
	if err != nil {
		t.Fatalf("Readdirnames failed: %v", err)
	}

	// Should use native implementation
	if len(names) != 2 || names[0] != "custom1.txt" || names[1] != "custom2.txt" {
		t.Errorf("expected custom names, got %v", names)
	}
}
