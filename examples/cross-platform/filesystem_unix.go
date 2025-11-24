//go:build !windows

package main

import (
	"github.com/absfs/absfs"
	"github.com/absfs/osfs"
)

// NewFS creates an OS filesystem.
// On Unix/macOS, no drive mapping is needed - Unix-style paths work natively.
// The drive parameter is ignored for cross-platform function signature compatibility.
func NewFS(drive string) absfs.FileSystem {
	fs, err := osfs.NewFS()
	if err != nil {
		panic(err) // In production, handle this error appropriately
	}
	return fs
}
