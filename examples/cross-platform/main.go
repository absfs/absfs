package main

import (
	"flag"
	"fmt"
	"io"
	"log"
)

func main() {
	drive := flag.String("drive", "", "Windows drive letter (e.g., 'D:'). Ignored on Unix/macOS. Defaults to 'C:' on Windows.")
	flag.Parse()

	// Create filesystem with platform-appropriate setup
	fs := NewFS(*drive)

	fmt.Println("Cross-Platform Filesystem Example")
	fmt.Println("==================================")
	fmt.Println()

	// Create a test directory structure
	testDir := "/tmp/absfs-example"
	fmt.Printf("Creating directory: %s\n", testDir)
	if err := fs.MkdirAll(testDir, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	// Create a file
	testFile := "/tmp/absfs-example/test.txt"
	fmt.Printf("Creating file: %s\n", testFile)
	f, err := fs.Create(testFile)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}

	// Write content
	content := "Hello from absfs! This works on Windows, macOS, and Linux.\n"
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		log.Fatalf("Failed to write to file: %v", err)
	}
	f.Close()

	// Read it back
	fmt.Printf("Reading file: %s\n", testFile)
	f, err = fs.Open(testFile)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	fmt.Println()
	fmt.Println("File contents:")
	fmt.Println(string(data))

	// Show file info
	info, err := fs.Stat(testFile)
	if err != nil {
		log.Fatalf("Failed to stat file: %v", err)
	}

	fmt.Printf("File size: %d bytes\n", info.Size())
	fmt.Printf("Modified: %s\n", info.ModTime())

	// Clean up
	fmt.Printf("\nCleaning up: removing %s\n", testDir)
	if err := fs.RemoveAll(testDir); err != nil {
		log.Fatalf("Failed to remove directory: %v", err)
	}

	fmt.Println("\n✅ Success! The same code worked on your platform.")
	fmt.Println()
	fmt.Println("On Windows: /tmp/absfs-example → C:\\tmp\\absfs-example (or your chosen drive)")
	fmt.Println("On Unix/macOS: /tmp/absfs-example → /tmp/absfs-example")
}
