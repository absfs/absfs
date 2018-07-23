package absfs

import (
	"io"
	"io/ioutil"

	"os"
)

type UnSeekable interface {

	// Name returns the name of the file as presented to Open.
	Name() string

	// Read reads up to len(b) bytes from the File. It returns the number of bytes
	// read and any error encountered. At end of file, Read returns 0, io.EOF.
	Read(p []byte) (int, error)

	// Write writes len(b) bytes to the File. It returns the number of bytes
	// written and an error, if any. Write returns a non-nil error when
	// n != len(b).
	Write(p []byte) (int, error)

	// Close closes the File, rendering it unusable for I/O. It returns an error,
	// if any.
	Close() error

	// Sync commits the current contents of the file to stable storage. Typically,
	// this means flushing the file system's in-memory copy of recently written
	// data to disk.
	Sync() error

	// Stat returns the FileInfo structure describing file. If there is an error,
	// it will be of type *PathError.
	Stat() (os.FileInfo, error)

	// Readdir reads the contents of the directory associated with file and
	// returns a slice of up to n FileInfo values, as would be returned by Lstat,
	// in directory order. Subsequent calls on the same file will yield further
	// FileInfos.

	// If n > 0, Readdir returns at most n FileInfo structures. In this case, if
	// Readdir returns an empty slice, it will return a non-nil error explaining
	// why. At the end of a directory, the error is io.EOF.

	// If n <= 0, Readdir returns all the FileInfo from the directory in a single
	// slice. In this case, if Readdir succeeds (reads all the way to the end of
	// the directory), it returns the slice and a nil error. If it encounters an
	// error before the end of the directory, Readdir returns the FileInfo read
	// until that point and a non-nil error.
	Readdir(int) ([]os.FileInfo, error)
}

type Seekable interface {
	UnSeekable

	// Seek sets the offset for the next Read or Write on file to offset,
	// interpreted according to whence: 0 means relative to the origin of the
	// file, 1 means relative to the current offset, and 2 means relative to the
	// end. It returns the new offset and an error, if any. The behavior of Seek
	// on a file opened with O_APPEND is not specified.
	Seek(offset int64, whence int) (ret int64, err error)
}

type File interface {
	Seekable

	// ReadAt reads len(b) bytes from the File starting at byte offset off. It
	// returns the number of bytes read and the error, if any. ReadAt always
	// returns a non-nil error when n < len(b). At end of file, that error is
	// io.EOF.
	ReadAt(b []byte, off int64) (n int, err error)

	// WriteAt writes len(b) bytes to the File starting at byte offset off. It
	// returns the number of bytes written and an error, if any. WriteAt returns
	// a non-nil error when n != len(b).
	WriteAt(b []byte, off int64) (n int, err error)

	// WriteString is like Write, but writes the contents of string s rather than
	// a slice of bytes.
	WriteString(s string) (n int, err error)

	// Truncate changes the size of the file. It does not change the I/O offset.
	// If there is an error, it will be of type *PathError.
	Truncate(size int64) error

	// If n > 0, Readdirnames returns at most n names. In this case, if
	// Readdirnames returns an empty slice, it will return a non-nil error
	// explaining why. At the end of a directory, the error is io.EOF.

	// If n <= 0, Readdirnames returns all the names from the directory in a single
	// slice. In this case, if Readdirnames succeeds (reads all the way to the end of
	// the directory), it returns the slice and a nil error. If it encounters an
	// error before the end of the directory, Readdirnames returns the names read
	// until that point and a non-nil error.
	Readdirnames(n int) (names []string, err error)
}

func ExtendUnseekable(uf UnSeekable) (Seekable, error) {
	data, err := ioutil.ReadAll(uf)
	if err != nil {
		return nil, err
	}
	sb := &seekbuffer{
		data:   data,
		offset: 0,
		uf:     uf,
	}
	return sb, nil
}

type InvalidFile struct{}

func (f *InvalidFile) Name() string {
	return ""
}

func (f *InvalidFile) Read(p []byte) (int, error) {
	return 0, os.ErrInvalid
}

func (f *InvalidFile) Write(p []byte) (int, error) {
	return 0, os.ErrInvalid
}

func (f *InvalidFile) Close() error {
	return os.ErrInvalid
}

func (f *InvalidFile) Sync() error {
	return os.ErrInvalid
}

func (f *InvalidFile) Stat() (os.FileInfo, error) {
	return nil, os.ErrInvalid
}

func (f *InvalidFile) Readdir(int) ([]os.FileInfo, error) {
	return nil, os.ErrInvalid
}

func (f *InvalidFile) Seek(offset int64, whence int) (ret int64, err error) {
	return 0, os.ErrInvalid
}

func (f *InvalidFile) ReadAt(b []byte, off int64) (n int, err error) {
	return 0, os.ErrInvalid
}

func (f *InvalidFile) WriteAt(b []byte, off int64) (n int, err error) {
	return 0, os.ErrInvalid
}

func (f *InvalidFile) WriteString(s string) (n int, err error) {
	return 0, os.ErrInvalid
}

func (f *InvalidFile) Truncate(size int64) error {
	return os.ErrInvalid
}

func (f *InvalidFile) Readdirnames(n int) (names []string, err error) {
	return nil, os.ErrInvalid
}

func ExtendSeekable(sf Seekable) File {
	return &file{sf}
}

type seekbuffer struct {
	data   []byte
	offset int64
	uf     UnSeekable
}

func (sb *seekbuffer) Name() string {
	return sb.uf.Name()
}

func (sb *seekbuffer) Read(p []byte) (int, error) {
	if sb.offset >= int64(len(sb.data)) {
		return 0, io.EOF
	}
	n := copy(p, sb.data[sb.offset:])
	sb.offset += int64(n)
	return n, nil
}

func (sb *seekbuffer) Write(p []byte) (int, error) {
	if int64(len(p))+sb.offset > int64(len(sb.data)) {
		data := make([]byte, int(int64(len(p))+sb.offset))
		copy(data, sb.data)
		sb.data = data
	}
	n := copy(sb.data[int(sb.offset):], p)
	return n, nil
}

func (sb *seekbuffer) Close() error {
	err := sb.Sync()
	if err != nil {
		return err
	}

	return sb.uf.Close()
}

func (sb *seekbuffer) Sync() error {
	n, err := sb.uf.Write(sb.data)
	if n != len(sb.data) {
		panic("something went wrong")
		return err
	}

	return sb.uf.Sync()
}

func (sb *seekbuffer) Stat() (os.FileInfo, error) {
	return sb.uf.Stat()
}

func (sb *seekbuffer) Readdir(n int) ([]os.FileInfo, error) {
	return sb.uf.Readdir(n)
}

func (sb *seekbuffer) Seek(offset int64, whence int) (ret int64, err error) {
	switch whence {
	case io.SeekStart:
		sb.offset = offset
	case io.SeekCurrent:
		sb.offset += offset
	case io.SeekEnd:
		sb.offset = int64(len(sb.data)) + offset
	}
	return sb.offset, nil
}

type file struct {
	sf Seekable
}

func (f *file) Name() string {
	return f.sf.Name()
}

func (f *file) Read(p []byte) (int, error) {
	return f.sf.Read(p)
}

func (f *file) Readdir(n int) ([]os.FileInfo, error) {
	return f.sf.Readdir(n)
}

// func (f *file) Readdir(n int) ([]os.FileInfo, error) {
// 	return f.sf.Readdir(n)
// }

func (f *file) Sync() error {
	return f.sf.Sync()
}

func (f *file) Stat() (os.FileInfo, error) {
	return f.sf.Stat()
}

func (f *file) Seek(offset int64, whence int) (ret int64, err error) {
	return f.sf.Seek(offset, whence)
}

func (f *file) ReadAt(b []byte, off int64) (n int, err error) {
	if file, ok := f.sf.(ater); ok {
		return file.ReadAt(b, off)
	}
	_, err = f.sf.Seek(off, io.SeekStart)
	if err != nil {
		return 0, err
	}
	return f.sf.Read(b)
}

func (f *file) Write(p []byte) (int, error) {
	return f.sf.Write(p)
}

func (f *file) WriteAt(b []byte, off int64) (n int, err error) {
	if file, ok := f.sf.(ater); ok {
		return file.WriteAt(b, off)
	}
	_, err = f.sf.Seek(off, io.SeekStart)
	if err != nil {
		return 0, err
	}
	return f.sf.Write(b)
}

func (f *file) WriteString(s string) (n int, err error) {
	if file, ok := f.sf.(stringwriter); ok {
		return file.WriteString(s)
	}

	return f.sf.Write([]byte(s))
}

func (f *file) Truncate(size int64) error {
	if ff, ok := f.sf.(filetruncater); ok {
		return ff.Truncate(size)
	}

	_, err := f.sf.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	bufsize := 512
	buf := make([]byte, 512)
	for i := 0; int64(i) < int64(size); i += 512 {
		if size-int64(i) < 512 {
			bufsize = int(size - int64(i))
		}

		_, err = f.sf.Write(buf[:bufsize])
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *file) Readdirnames(n int) (names []string, err error) {
	if file, ok := f.sf.(dirnamer); ok {
		return file.Readdirnames(n)
	}
	var infos []os.FileInfo
	infos, err = f.sf.Readdir(n)
	if err != nil {
		return nil, err
	}

	names = make([]string, len(infos))

	for i, info := range infos {
		names[i] = info.Name()
	}

	return names, nil
}

func (f *file) Close() error {
	return nil
}

// interfaces for easy method typing

type ater interface {
	ReadAt(b []byte, off int64) (n int, err error)
	WriteAt(b []byte, off int64) (n int, err error)
}

type stringwriter interface {
	WriteString(s string) (n int, err error)
}

type filetruncater interface {
	Truncate(size int64) error
}

type dirnamer interface {
	Readdirnames(n int) (names []string, err error)
}
