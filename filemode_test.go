package absfs

import (
	"os"
	"testing"
)

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
