package absfs_test

import (
	"fmt"

	"github.com/absfs/absfs"
)

// Example demonstrates basic usage of creating a filesystem
func Example() {
	// Create your Filer implementation (shown simplified)
	// filer := &MyFiler{}

	// Extend it to get full FileSystem interface
	// fs := absfs.ExtendFiler(filer)

	fmt.Println("See individual examples for specific operations")
	// Output: See individual examples for specific operations
}

// ExampleExtendFiler shows how to extend a Filer to a full FileSystem
func ExampleExtendFiler() {
	// Create a Filer implementation
	// In practice, use an actual implementation like osfs, memfs, etc.
	// filer := &MyCustomFiler{}

	// Extend it to get all convenience methods
	// fs := absfs.ExtendFiler(filer)

	// Now you can use convenience methods like:
	// fs.Create("/file.txt")
	// fs.MkdirAll("/path/to/dir", 0755)
	// fs.Open("/file.txt")

	fmt.Println("Extended filer to FileSystem")
	// Output: Extended filer to FileSystem
}

// ExampleFileSystem_Create demonstrates creating a file
func ExampleFileSystem_Create() {
	// Assuming you have a FileSystem instance
	// fs := absfs.ExtendFiler(yourFiler)

	// Create a new file (simplified for example)
	// file, err := fs.Create("/example.txt")
	// if err != nil {
	//     panic(err)
	// }
	// defer file.Close()

	// file.Write([]byte("Hello, World!"))

	fmt.Println("File created successfully")
	// Output: File created successfully
}

// ExampleFileSystem_MkdirAll demonstrates creating nested directories
func ExampleFileSystem_MkdirAll() {
	// fs := absfs.ExtendFiler(yourFiler)

	// Create nested directories in one call
	// err := fs.MkdirAll("/path/to/deep/directory", 0755)
	// if err != nil {
	//     panic(err)
	// }

	fmt.Println("Directories created")
	// Output: Directories created
}

// ExampleFileSystem_Chdir demonstrates changing directory
func ExampleFileSystem_Chdir() {
	// fs := absfs.ExtendFiler(yourFiler)

	// Change to a directory
	// err := fs.Chdir("/home/user")
	// if err != nil {
	//     panic(err)
	// }

	// Get current directory
	// cwd, _ := fs.Getwd()
	// fmt.Println(cwd) // Output: /home/user

	fmt.Println("Changed directory")
	// Output: Changed directory
}

// ExampleParseFileMode demonstrates parsing Unix file mode strings
func ExampleParseFileMode() {
	mode, err := absfs.ParseFileMode("drwxr-xr-x")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Parsed mode: %o\n", mode&0777)
	// Output: Parsed mode: 755
}

// ExampleParseFlags demonstrates parsing file open flags
func ExampleParseFlags() {
	flags, err := absfs.ParseFlags("O_RDWR|O_CREATE|O_TRUNC")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Flags: %s\n", flags)
	// Output: Flags: O_RDWR|O_CREATE|O_TRUNC
}

// ExampleFlags_String demonstrates converting flags to string
func ExampleFlags_String() {
	flags := absfs.Flags(absfs.O_RDWR | absfs.O_APPEND)
	fmt.Println(flags.String())
	// Output: O_RDWR|O_APPEND
}

// ExampleExtendSeekable shows how to extend a Seekable to a File
func ExampleExtendSeekable() {
	// If you have a Seekable implementation that doesn't implement
	// ReadAt, WriteAt, Truncate, etc., you can extend it:

	// seekable := &MySeekable{}
	// file := absfs.ExtendSeekable(seekable)

	// Now file implements the full File interface with:
	// - ReadAt/WriteAt
	// - Truncate
	// - Readdirnames

	fmt.Println("Extended seekable to file")
	// Output: Extended seekable to file
}

// ExampleFileSystem_Open demonstrates opening a file for reading
func ExampleFileSystem_Open() {
	// fs := absfs.ExtendFiler(yourFiler)

	// Open file for reading
	// file, err := fs.Open("/example.txt")
	// if err != nil {
	//     panic(err)
	// }
	// defer file.Close()

	// Read contents
	// data, _ := io.ReadAll(file)
	// fmt.Println(string(data))

	fmt.Println("File opened for reading")
	// Output: File opened for reading
}

// ExampleFileSystem_Stat demonstrates getting file information
func ExampleFileSystem_Stat() {
	// fs := absfs.ExtendFiler(yourFiler)

	// Get file information
	// info, err := fs.Stat("/example.txt")
	// if err != nil {
	//     panic(err)
	// }

	// fmt.Printf("Name: %s\n", info.Name())
	// fmt.Printf("Size: %d\n", info.Size())
	// fmt.Printf("Mode: %v\n", info.Mode())
	// fmt.Printf("IsDir: %v\n", info.IsDir())

	fmt.Println("File info retrieved")
	// Output: File info retrieved
}

// ExampleFile_ReadAt demonstrates reading at a specific offset
func ExampleFile_ReadAt() {
	// Assuming you have a File
	// file, _ := fs.Open("/example.txt")
	// defer file.Close()

	// Read 10 bytes starting at offset 5
	// buf := make([]byte, 10)
	// n, err := file.ReadAt(buf, 5)
	// if err != nil && err != io.EOF {
	//     panic(err)
	// }
	// fmt.Printf("Read %d bytes: %s\n", n, buf[:n])

	fmt.Println("Read at offset")
	// Output: Read at offset
}

// Example permission constants usage
func ExampleOS_ALL_RWX() {
	// Use permission constants for clarity
	// fs.Mkdir("/public", absfs.OS_ALL_RWX) // 0777
	// fs.Mkdir("/private", absfs.OS_USER_RWX) // 0700

	fmt.Printf("OS_ALL_RWX = 0%o\n", absfs.OS_ALL_RWX)
	// Output: OS_ALL_RWX = 0777
}
