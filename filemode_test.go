package absfs

import (
	"os"
	"testing"
)

func TestPermissionConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant os.FileMode
		expected os.FileMode
	}{
		{"OS_USER_R", OS_USER_R, 0400},
		{"OS_USER_W", OS_USER_W, 0200},
		{"OS_USER_X", OS_USER_X, 0100},
		{"OS_USER_RW", OS_USER_RW, 0600},
		{"OS_USER_RWX", OS_USER_RWX, 0700},
		{"OS_GROUP_R", OS_GROUP_R, 0040},
		{"OS_GROUP_W", OS_GROUP_W, 0020},
		{"OS_GROUP_X", OS_GROUP_X, 0010},
		{"OS_GROUP_RW", OS_GROUP_RW, 0060},
		{"OS_GROUP_RWX", OS_GROUP_RWX, 0070},
		{"OS_OTH_R", OS_OTH_R, 0004},
		{"OS_OTH_W", OS_OTH_W, 0002},
		{"OS_OTH_X", OS_OTH_X, 0001},
		{"OS_OTH_RW", OS_OTH_RW, 0006},
		{"OS_OTH_RWX", OS_OTH_RWX, 0007},
		{"OS_ALL_R", OS_ALL_R, 0444},
		{"OS_ALL_W", OS_ALL_W, 0222},
		{"OS_ALL_X", OS_ALL_X, 0111},
		{"OS_ALL_RW", OS_ALL_RW, 0666},
		{"OS_ALL_RWX", OS_ALL_RWX, 0777}, // This is the critical bug fix verification
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.constant != test.expected {
				t.Errorf("%s = %#o, expected %#o", test.name, test.constant, test.expected)
			}
		})
	}
}

func TestParseFileModeErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"too short", "drwx"},
		{"invalid type char", "Xrwxrwxrwx"},
		{"invalid permission char", "draxrwxrwx"},
		{"wrong char position", "drwxrwxrww"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseFileMode(test.input)
			if err == nil {
				t.Errorf("ParseFileMode(%q) expected error, got nil", test.input)
			}
		})
	}
}

func TestParseFileMode(t *testing.T) {
	// validtypes := "-acdDglLpSTtu"
	modes := []struct {
		C string
		M os.FileMode
	}{
		{C: "d", M: os.ModeDir},        // d: is a directory
		{C: "a", M: os.ModeAppend},     // a: append-only
		{C: "l", M: os.ModeExclusive},  // l: exclusive use
		{C: "T", M: os.ModeTemporary},  // T: temporary file; Plan 9 only
		{C: "L", M: os.ModeSymlink},    // L: symbolic link
		{C: "D", M: os.ModeDevice},     // D: device file
		{C: "p", M: os.ModeNamedPipe},  // p: named pipe (FIFO)
		{C: "S", M: os.ModeSocket},     // S: Unix domain socket
		{C: "u", M: os.ModeSetuid},     // u: setuid
		{C: "g", M: os.ModeSetgid},     // g: setgid
		{C: "c", M: os.ModeCharDevice}, // c: Unix character device, when ModeDevice is set
		{C: "t", M: os.ModeSticky},     // t: sticky
	}

	usrPerm := []os.FileMode{0, OS_USER_R, OS_USER_W, OS_USER_X,
		OS_USER_R | OS_USER_W, OS_USER_R | OS_USER_X, OS_USER_W | OS_USER_X, OS_USER_RWX}
	grpPerm := []os.FileMode{0, OS_GROUP_R, OS_GROUP_W, OS_GROUP_X,
		OS_GROUP_R | OS_GROUP_W, OS_GROUP_R | OS_GROUP_X, OS_GROUP_W | OS_GROUP_X,
		OS_GROUP_RWX}
	otrPerm := []os.FileMode{0, OS_OTH_R, OS_OTH_W, OS_OTH_X, OS_OTH_R | OS_OTH_W,
		OS_OTH_R | OS_OTH_X, OS_OTH_W | OS_OTH_X, OS_OTH_RWX}

	type testcase struct {
		In  string
		Exp os.FileMode
	}

	var tests []testcase
	for _, mode := range modes {
		for _, usr := range usrPerm {
			for _, grp := range grpPerm {
				for _, oth := range otrPerm {
					m := mode.M | usr | grp | oth
					tests = append(tests, testcase{In: m.String(), Exp: m})
				}
			}
		}
	}

	for _, test := range tests {
		m, err := ParseFileMode(test.In)
		if err != nil {
			t.Errorf("%s %o, %o, %q", err, m, test.Exp, test.In)
			continue
		}
		if m != test.Exp {
			t.Errorf("got %o, expected %o from %q", m, test.Exp, test.In)
		}
	}
}
