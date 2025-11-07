package absfs

import (
	"os"
	"testing"
)

func BenchmarkExtendFiler(b *testing.B) {
	mock := newMockFiler()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ExtendFiler(mock)
	}
}

func BenchmarkFileSystemCreate(b *testing.B) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f, _ := fs.Create("/bench.txt")
		f.Close()
		fs.Remove("/bench.txt")
	}
}

func BenchmarkFileSystemMkdir(b *testing.B) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.Mkdir("/benchdir", 0755)
		fs.Remove("/benchdir")
	}
}

func BenchmarkFileSystemMkdirAll(b *testing.B) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.MkdirAll("/a/b/c/d", 0755)
		fs.Remove("/a")
	}
}

func BenchmarkFileSystemStat(b *testing.B) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)
	fs.Create("/bench.txt")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.Stat("/bench.txt")
	}
}

func BenchmarkFileSystemChdir(b *testing.B) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)
	fs.Mkdir("/benchdir", 0755)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.Chdir("/benchdir")
		fs.Chdir("/")
	}
}

func BenchmarkFileSystemRelativePathResolution(b *testing.B) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)
	fs.Mkdir("/work", 0755)
	fs.Chdir("/work")
	fs.Create("/work/file.txt")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.Open("file.txt")
	}
}

func BenchmarkParseFileMode(b *testing.B) {
	input := "drwxr-xr-x"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseFileMode(input)
	}
}

func BenchmarkParseFlags(b *testing.B) {
	input := "O_RDWR|O_CREATE|O_TRUNC"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseFlags(input)
	}
}

func BenchmarkFlagsString(b *testing.B) {
	flags := Flags(O_RDWR | O_CREATE | O_APPEND)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = flags.String()
	}
}

func BenchmarkExtendSeekable(b *testing.B) {
	ts := &testSeekable{name: "test.txt", data: []byte("hello world")}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ExtendSeekable(ts)
	}
}

func BenchmarkFileRead(b *testing.B) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)
	f, _ := fs.Create("/bench.txt")
	f.Write([]byte("hello world benchmark data"))
	f.Close()

	buf := make([]byte, 26)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f, _ := fs.Open("/bench.txt")
		f.Read(buf)
		f.Close()
	}
}

func BenchmarkFileWrite(b *testing.B) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)
	data := []byte("benchmark write data")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f, _ := fs.Create("/bench.txt")
		f.Write(data)
		f.Close()
		fs.Remove("/bench.txt")
	}
}

func BenchmarkFileSeek(b *testing.B) {
	mock := newMockFiler()
	fs := ExtendFiler(mock)
	f, _ := fs.Create("/bench.txt")
	f.Write([]byte("0123456789abcdefghijklmnopqrstuvwxyz"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Seek(0, 0)
		f.Seek(10, 0)
		f.Seek(0, 2)
	}
	f.Close()
}

func BenchmarkPermissionConstants(b *testing.B) {
	b.Run("OS_ALL_RWX", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = os.FileMode(OS_ALL_RWX)
		}
	})
	b.Run("OS_USER_RWX", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = os.FileMode(OS_USER_RWX)
		}
	})
}
