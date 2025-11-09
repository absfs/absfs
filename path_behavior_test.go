package absfs

import (
	"path/filepath"
	"runtime"
	"testing"
)

// TestPathBehaviorDocumentation documents how different path types are handled
// across platforms. This serves as both a test and documentation.
func TestPathBehaviorDocumentation(t *testing.T) {
	tests := []struct {
		path        string
		description string
	}{
		{"/unix/style", "Unix absolute path"},
		{"\\windows\\style", "Windows backslash path"},
		{"C:\\windows\\drive", "Windows drive letter path"},
		{"//server/share", "UNC-style with forward slashes"},
		{"\\\\server\\share", "UNC-style with backslashes"},
		{"relative/path", "Relative path"},
		{"/", "Root"},
		{"\\", "Backslash root"},
	}

	t.Logf("Platform: %s", runtime.GOOS)
	t.Logf("Native separator: %c", filepath.Separator)
	t.Log("\nPath behavior analysis:")
	t.Log("Path                      | IsAbs | isVirtualAbs | Description")
	t.Log("--------------------------|-------|--------------|------------------")

	for _, tt := range tests {
		osAbs := filepath.IsAbs(tt.path)
		virtAbs := isVirtualAbs(tt.path)
		t.Logf("%-25s | %-5t | %-12t | %s",
			tt.path, osAbs, virtAbs, tt.description)
	}
}
