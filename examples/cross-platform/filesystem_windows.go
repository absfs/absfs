//go:build windows

package main

import (
	"github.com/absfs/absfs"
	"github.com/absfs/osfs"
)

// NewFS creates an OS filesystem with drive mapping on Windows.
// This allows Unix-style absolute paths like "/config/app.json" to map to "C:\config\app.json".
func NewFS(drive string) absfs.FileSystem {
	if drive == "" {
		drive = "C:" // Default to C: drive
	}
	base, err := osfs.NewFS()
	if err != nil {
		panic(err) // In production, handle this error appropriately
	}
	return osfs.NewWindowsDriveMapper(base, drive)
}
