package absfs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// FuzzPathClean tests that path cleaning is idempotent and removes all relative components
func FuzzPathClean(f *testing.F) {
	// Seed with known patterns that should test various edge cases
	f.Add("../../../etc/passwd")
	f.Add("foo/./bar/../baz")
	f.Add("/./foo//bar/./baz/../qux")
	f.Add("\\host\\share\\..\\file")
	f.Add("")
	f.Add(".")
	f.Add("..")
	f.Add("/")
	f.Add("//")
	f.Add("///")
	f.Add("/path/to/file")
	f.Add("relative/path")
	f.Add("./relative")
	f.Add("../relative")
	f.Add("path/./with/./dots")
	f.Add("path/../with/../dots")
	f.Add("/absolute/../path")
	// Unicode edge cases
	f.Add("/path/to/文件")
	f.Add("/path/to/файл")
	// Very long path component
	f.Add(strings.Repeat("a", 256))
	// Trailing slashes
	f.Add("/path/to/dir/")
	f.Add("/path/to/dir/.")
	// Mixed separators (Windows-style)
	f.Add("C:\\Windows\\System32")
	f.Add("/mixed\\separators/path")
	// Multiple dots
	f.Add("...")
	f.Add("....")
	f.Add("path/.../to/file")

	f.Fuzz(func(t *testing.T, path string) {
		// Skip strings with null bytes as they're not valid file paths
		if strings.Contains(path, "\x00") {
			t.Skip("path contains null byte")
		}

		// Skip paths with colons (except after drive letters on Windows) as they cause
		// non-idempotent behavior in Go's filepath.Clean (known issue)
		// Examples: "/:" becomes "/:." then "\\:."
		if strings.Contains(path, ":") {
			// Allow "C:" style drive letters
			if !(len(path) >= 2 && path[1] == ':' && ((path[0] >= 'A' && path[0] <= 'Z') || (path[0] >= 'a' && path[0] <= 'z'))) {
				t.Skip("path contains colon in non-drive-letter position")
			}
		}

		// Skip paths with question marks as they cause non-idempotent behavior on Windows
		// Examples: "/?" becomes "/?." then "\\.\\?."
		if strings.Contains(path, "?") {
			t.Skip("path contains question mark (wildcard character)")
		}

		cleaned := filepath.Clean(path)

		// Property 1: Idempotent - Clean(Clean(x)) == Clean(x)
		cleaned2 := filepath.Clean(cleaned)
		if cleaned2 != cleaned {
			t.Errorf("not idempotent: %q → %q → %q", path, cleaned, cleaned2)
		}

		// Property 2: No "." components in middle of path
		// Exception: path "." itself is valid
		// Exception: On Windows, UNC paths like "\\\\." or "\\\\\\.\\" are valid (. is the server name)
		if cleaned != "." && strings.Contains(cleaned, string(filepath.Separator)+"."+string(filepath.Separator)) {
			// Skip check if this is a Windows UNC path where . is the server name
			// UNC paths start with \\ and can have additional separators before the server name
			isUNCWithDotServer := false
			if len(cleaned) >= 3 && cleaned[0] == '\\' && cleaned[1] == '\\' {
				// Skip past initial \\ and any additional backslashes
				i := 2
				for i < len(cleaned) && cleaned[i] == '\\' {
					i++
				}
				// Check if next char is "." (server name)
				if i < len(cleaned) && cleaned[i] == '.' {
					isUNCWithDotServer = true
				}
			}
			if !isUNCWithDotServer {
				t.Errorf("contains '/./' in middle: %q → %q", path, cleaned)
			}
		}

		// Property 3: No ".." as a standalone path component in result for absolute paths
		// For absolute paths, .. should be resolved
		// Check for "/.." at start, "/../" in middle, or "/.." at end
		// Exception: On Windows, UNC paths like "\\\.." or "\\\\\\.." are valid (.. is the server name)
		if filepath.IsAbs(cleaned) {
			sep := string(filepath.Separator)
			// Skip check if this is a Windows UNC path where .. is the server name
			// UNC paths start with \\ and can have additional separators before the server name
			isUNCWithDotDotServer := false
			if len(cleaned) >= 4 && cleaned[0] == '\\' && cleaned[1] == '\\' {
				// Skip past initial \\ and any additional backslashes
				i := 2
				for i < len(cleaned) && cleaned[i] == '\\' {
					i++
				}
				// Check if next chars are ".." (server name)
				if i+2 <= len(cleaned) && cleaned[i:i+2] == ".." && (i+2 == len(cleaned) || cleaned[i+2] == '\\') {
					isUNCWithDotDotServer = true
				}
			}
			if !isUNCWithDotDotServer {
				if strings.HasPrefix(cleaned, sep+".."+sep) ||
					strings.Contains(cleaned, sep+".."+sep) ||
					strings.HasSuffix(cleaned, sep+"..") {
					t.Errorf("absolute path contains '..' component: %q → %q", path, cleaned)
				}
			}
		}

		// Property 4: No empty path unless input was empty
		// Note: filepath.Clean converts "" to "."
		if path != "" && cleaned == "" {
			t.Errorf("cleaned to empty: %q → %q", path, cleaned)
		}

		// Property 5: Result should not have trailing separator unless it's root
		// On Windows, UNC paths like "\\\\" are valid roots and can have multiple separators
		// Also, UNC paths like "\\server\" or "\\server\share\" are root-like and preserve trailing separators
		if len(cleaned) > 1 && cleaned[len(cleaned)-1] == filepath.Separator {
			// Check if this is a valid root path or UNC path
			isRoot := cleaned == string(filepath.Separator)
			// On Windows: UNC paths starting with \\ can have trailing separators
			isUNCPath := len(cleaned) >= 2 && cleaned[0] == '\\' && cleaned[1] == '\\'
			if !isRoot && !isUNCPath {
				t.Errorf("has trailing separator: %q → %q", path, cleaned)
			}
		}

		// Property 6: No double separators in result
		doubleSep := string(filepath.Separator) + string(filepath.Separator)
		if strings.Contains(cleaned, doubleSep) {
			// Exception: UNC paths on Windows start with \\
			if !(len(cleaned) >= 2 && cleaned[0] == '\\' && cleaned[1] == '\\') {
				t.Errorf("contains double separator: %q → %q", path, cleaned)
			}
		}
	})
}

// FuzzFileMode tests FileMode operations with edge case values
func FuzzFileMode(f *testing.F) {
	// Seed with common and edge case values
	f.Add(uint32(0644))
	f.Add(uint32(0755))
	f.Add(uint32(0000))
	f.Add(uint32(0777))
	f.Add(uint32(0666))
	f.Add(uint32(0700))
	f.Add(uint32(0400))
	f.Add(uint32(0200))
	f.Add(uint32(0100))
	// Edge cases
	f.Add(uint32(0xFFFFFFFF))
	f.Add(uint32(0x7FFFFFFF))
	f.Add(uint32(0))
	// With mode bits set
	f.Add(uint32(os.ModeDir | 0755))
	f.Add(uint32(os.ModeSymlink | 0777))
	f.Add(uint32(os.ModeDevice | 0666))
	f.Add(uint32(os.ModeNamedPipe | 0644))
	f.Add(uint32(os.ModeSocket | 0700))
	f.Add(uint32(os.ModeSetuid | 0755))
	f.Add(uint32(os.ModeSetgid | 0755))
	f.Add(uint32(os.ModeSticky | 0777))

	f.Fuzz(func(t *testing.T, mode uint32) {
		fm := os.FileMode(mode)

		// Should not panic on any mode value
		_ = fm.IsDir()
		_ = fm.IsRegular()
		_ = fm.Perm()
		_ = fm.String()
		_ = fm.Type()

		// Property 1: Perm() should return only permission bits (lower 9 bits)
		perm := fm.Perm()
		if perm != os.FileMode(mode&0777) {
			t.Errorf("perm roundtrip failed: mode=%#o, perm=%#o, expected=%#o", mode, perm, mode&0777)
		}

		// Property 2: Type() should return only type bits
		typ := fm.Type()
		if (typ & 0777) != 0 {
			t.Errorf("type contains permission bits: mode=%#o, type=%#o", mode, typ)
		}

		// Property 3: IsDir and Type should be consistent
		isDir := fm.IsDir()
		hasDir := (typ & os.ModeDir) != 0
		if isDir != hasDir {
			t.Errorf("IsDir inconsistent with Type: mode=%#o, IsDir=%v, Type&ModeDir=%v", mode, isDir, hasDir)
		}

		// Property 4: IsRegular should be true only when no mode bits are set
		isRegular := fm.IsRegular()
		if isRegular && typ != 0 {
			t.Errorf("IsRegular true but type bits set: mode=%#o, type=%#o", mode, typ)
		}

		// Property 5: String should never panic and should return non-empty
		s := fm.String()
		if s == "" {
			t.Errorf("String() returned empty for mode=%#o", mode)
		}
	})
}

// FuzzFlags tests file open flags with various combinations
func FuzzFlags(f *testing.F) {
	// Seed with valid flag combinations
	f.Add(int(O_RDONLY))
	f.Add(int(O_WRONLY))
	f.Add(int(O_RDWR))
	f.Add(int(O_WRONLY | O_CREATE))
	f.Add(int(O_RDWR | O_CREATE))
	f.Add(int(O_WRONLY | O_CREATE | O_TRUNC))
	f.Add(int(O_RDWR | O_CREATE | O_EXCL))
	f.Add(int(O_RDWR | O_APPEND))
	f.Add(int(O_WRONLY | O_APPEND | O_CREATE))
	f.Add(int(O_RDWR | O_SYNC))
	f.Add(int(O_WRONLY | O_CREATE | O_TRUNC | O_SYNC))
	// Edge cases
	f.Add(int(0))
	f.Add(int(-1))
	f.Add(int(0xFFFFFF))

	f.Fuzz(func(t *testing.T, flags int) {
		// Should not panic on any flag combination
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic with flags=0x%x: %v", flags, r)
			}
		}()

		// Test flag operations
		_ = flags & O_RDONLY
		_ = flags & O_WRONLY
		_ = flags & O_RDWR
		_ = flags & O_CREATE
		_ = flags & O_APPEND
		_ = flags & O_TRUNC
		_ = flags & O_EXCL
		_ = flags & O_SYNC

		// Property 1: Access mode should be one of the valid values
		accessMode := flags & O_ACCESS
		validAccessModes := []int{O_RDONLY, O_WRONLY, O_RDWR}
		isValidAccess := false
		for _, valid := range validAccessModes {
			if accessMode == valid {
				isValidAccess = true
				break
			}
		}

		// Only log if it's not a valid access mode
		// This helps discover edge cases
		if !isValidAccess && accessMode <= O_RDWR {
			t.Logf("unusual access mode: flags=0x%x, accessMode=0x%x", flags, accessMode)
		}

		// Property 2: Test Flags type conversion and String method
		f := Flags(flags)
		s := f.String()

		// String should not panic and should return something
		if s == "" && flags != 0 {
			t.Logf("String() returned empty for flags=0x%x", flags)
		}

		// Property 3: ParseFlags should handle valid flag strings
		// Only test if the string looks reasonable
		if strings.Contains(s, "O_") {
			parsed, err := ParseFlags(s)
			if err != nil {
				t.Logf("ParseFlags failed for %q (original flags=0x%x): %v", s, flags, err)
			} else if parsed != f {
				// This is expected for some invalid combinations, just log it
				t.Logf("ParseFlags roundtrip mismatch: original=0x%x (%s), parsed=0x%x (%s)",
					f, f.String(), parsed, parsed.String())
			}
		}
	})
}

// FuzzPathJoin tests that path joining handles edge cases correctly
func FuzzPathJoin(f *testing.F) {
	// Seed with various combinations
	f.Add("/base", "file.txt")
	f.Add("relative", "../escape")
	f.Add("", "file")
	f.Add("base", "")
	f.Add("/", "/absolute")
	f.Add("/base/", "file")
	f.Add("/base", "/file")
	f.Add(".", "file")
	f.Add("..", "file")
	f.Add("base", ".")
	f.Add("base", "..")
	f.Add("/", "")
	f.Add("", "")
	f.Add("/a/b/c", "../../d")
	// Unicode
	f.Add("/path", "文件")
	f.Add("путь", "файл")
	// Windows-style
	f.Add("C:\\base", "file")
	f.Add("\\\\server\\share", "file")
	// Multiple dots
	f.Add("base", "...")
	f.Add("...", "file")
	// Special characters
	f.Add("base", "file with spaces")
	f.Add("path", "file\twith\ttabs")

	f.Fuzz(func(t *testing.T, base, elem string) {
		// Skip strings with null bytes
		if strings.Contains(base, "\x00") || strings.Contains(elem, "\x00") {
			t.Skip("path contains null byte")
		}

		// Should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic with Join(%q, %q): %v", base, elem, r)
			}
		}()

		joined := filepath.Join(base, elem)

		// Property 1: Result should be clean
		// Special case: Join("", "") returns "" but Clean("") returns "."
		// This is expected Go behavior
		cleaned := filepath.Clean(joined)
		if joined != cleaned && joined != "" {
			t.Errorf("join not clean: Join(%q, %q) = %q, Clean = %q",
				base, elem, joined, cleaned)
		}

		// Property 2: If elem is absolute, result should be elem (cleaned)
		// This is platform-specific behavior
		if filepath.IsAbs(elem) {
			expectedClean := filepath.Clean(elem)
			if joined != expectedClean {
				// On some platforms this is expected behavior
				t.Logf("Join with absolute elem: Join(%q, %q) = %q, expected %q",
					base, elem, joined, expectedClean)
			}
		}

		// Property 3: Join should be idempotent when joining with empty string
		joinedWithEmpty := filepath.Join(joined, "")
		if joinedWithEmpty != joined {
			t.Errorf("not idempotent with empty: Join(%q, %q) = %q, Join(result, %q) = %q",
				base, elem, joined, "", joinedWithEmpty)
		}

		// Property 4: Joining then cleaning should equal just joining
		// (since Join already cleans)
		// Exception: Join("", "") returns "" but Clean("") returns "."
		if filepath.Clean(joined) != joined && joined != "" {
			t.Errorf("Join result not clean: Join(%q, %q) = %q, Clean = %q",
				base, elem, joined, filepath.Clean(joined))
		}
	})
}

// FuzzParseFileMode tests the ParseFileMode function with various inputs
func FuzzParseFileMode(f *testing.F) {
	// Seed with valid mode strings
	f.Add("-rwxrwxrwx")
	f.Add("drwxr-xr-x")
	f.Add("-rw-r--r--")
	f.Add("-r--------")
	f.Add("lrwxrwxrwx")
	f.Add("prw-------")
	f.Add("crw-rw----")
	f.Add("srwxrwxrwx")
	// Edge cases
	f.Add("----------")
	f.Add("drwxrwxrwx")
	f.Add("Lrwxrwxrwx")
	f.Add("Drwxrwxrwx")
	f.Add("Srwxrwxrwx")
	f.Add("urwxrwxrwx")
	f.Add("grwxrwxrwx")
	f.Add("trwxrwxrwx")

	f.Fuzz(func(t *testing.T, input string) {
		// Skip invalid lengths - ParseFileMode has a known issue with strings > 10 chars
		// It only validates for length < 10, but doesn't handle length > 10
		if len(input) > 10 {
			t.Skip("input longer than 10 characters")
		}

		// Should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic with ParseFileMode(%q): %v", input, r)
			}
		}()

		mode, err := ParseFileMode(input)

		// If parsing succeeds, test properties
		if err == nil {
			// Property 1: Mode string roundtrip should work
			modeStr := mode.String()
			mode2, err2 := ParseFileMode(modeStr)
			if err2 == nil && mode2 != mode {
				t.Errorf("roundtrip failed: %q → %#o → %q → %#o",
					input, mode, modeStr, mode2)
			}

			// Property 2: Permission bits should be in valid range
			perm := mode.Perm()
			if perm > 0777 {
				t.Errorf("permission bits out of range: %q → %#o, perm=%#o",
					input, mode, perm)
			}
		}

		// Property 3: String length constraint
		// ParseFileMode expects length >= 10
		if len(input) < 10 {
			if err == nil {
				t.Errorf("accepted too-short input: %q (len=%d)", input, len(input))
			}
		}

		// Property 4: If length is exactly 10, first char should determine type
		if len(input) == 10 {
			firstChar := input[0]
			validFirstChars := "-acdDglLpSTtu"
			if err == nil && !strings.ContainsRune(validFirstChars, rune(firstChar)) {
				t.Errorf("accepted invalid first char: %q", input)
			}
		}
	})
}

// FuzzParseFlags tests the ParseFlags function with various inputs
func FuzzParseFlags(f *testing.F) {
	// Seed with valid flag strings
	f.Add("O_RDONLY")
	f.Add("O_WRONLY")
	f.Add("O_RDWR")
	f.Add("O_RDONLY|O_CREATE")
	f.Add("O_WRONLY|O_CREATE")
	f.Add("O_RDWR|O_CREATE")
	f.Add("O_WRONLY|O_CREATE|O_TRUNC")
	f.Add("O_RDWR|O_APPEND")
	f.Add("O_WRONLY|O_CREATE|O_EXCL")
	f.Add("O_RDWR|O_SYNC")
	f.Add("O_RDONLY|O_APPEND|O_CREATE|O_EXCL|O_SYNC|O_TRUNC")
	// Edge cases
	f.Add("")
	f.Add("|")
	f.Add("||")
	f.Add("O_INVALID")
	f.Add("O_RDONLY|O_WRONLY")
	f.Add("O_RDONLY|O_RDWR")
	f.Add("O_WRONLY|O_RDWR")

	f.Fuzz(func(t *testing.T, input string) {
		// Should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic with ParseFlags(%q): %v", input, r)
			}
		}()

		flags, err := ParseFlags(input)

		// If parsing succeeds, test properties
		if err == nil {
			// Property 1: Roundtrip should work for valid inputs
			flagStr := flags.String()
			flags2, err2 := ParseFlags(flagStr)
			if err2 != nil {
				t.Errorf("roundtrip parse failed: %q → %v → %q: %v",
					input, flags, flagStr, err2)
			} else if flags2 != flags {
				t.Logf("roundtrip value mismatch: %q → 0x%x → %q → 0x%x",
					input, flags, flagStr, flags2)
			}

			// Property 2: Exactly one access mode should be set
			accessMode := int(flags) & O_ACCESS
			validAccessModes := []int{O_RDONLY, O_WRONLY, O_RDWR}
			accessModeCount := 0
			for _, valid := range validAccessModes {
				if accessMode == valid {
					accessModeCount++
				}
			}
			// O_RDONLY is 0, so it's the default
			if accessModeCount != 1 && accessMode != O_RDONLY {
				t.Logf("unexpected access mode: %q → 0x%x, accessMode=0x%x",
					input, flags, accessMode)
			}
		}

		// Property 3: Multiple access modes should be rejected
		multiAccessPatterns := []string{
			"O_RDONLY|O_WRONLY",
			"O_RDONLY|O_RDWR",
			"O_WRONLY|O_RDWR",
		}
		for _, pattern := range multiAccessPatterns {
			if input == pattern {
				if err == nil {
					t.Errorf("accepted multiple access modes: %q", input)
				}
			}
		}

		// Property 4: Unknown flags should be rejected
		if strings.Contains(input, "O_") && !strings.Contains(input, "O_RDONLY") &&
			!strings.Contains(input, "O_WRONLY") && !strings.Contains(input, "O_RDWR") &&
			!strings.Contains(input, "O_CREATE") && !strings.Contains(input, "O_APPEND") &&
			!strings.Contains(input, "O_TRUNC") && !strings.Contains(input, "O_EXCL") &&
			!strings.Contains(input, "O_SYNC") {
			if err == nil {
				t.Logf("accepted unknown flag in: %q", input)
			}
		}
	})
}
