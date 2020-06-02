package absfs

import (
	"strings"
	"testing"
)

type testType struct {
	In  Flags
	Exp string
}

func TestFlags(t *testing.T) {
	var a []testType
	flagStrings := []string{"O_APPEND", "O_CREATE", "O_EXCL", "O_SYNC", "O_TRUNC"}
	for i, flag := range []Flags{Flags(O_APPEND), Flags(O_CREATE), Flags(O_EXCL), Flags(O_SYNC), Flags(O_TRUNC)} {
		a = append(a, testType{flag, flagStrings[i]})
	}
	combinations := combinationsOf(a)

	c := 0
	modeStrings := []string{"O_RDONLY", "O_WRONLY", "O_RDWR"}
	for i, acc := range []Flags{Flags(O_RDONLY), Flags(O_WRONLY), Flags(O_RDWR)} {
		for _, test := range combinations {
			flags := Flags(acc) | test.In
			exp := modeStrings[i]
			if test.Exp != "" {
				exp = strings.Join([]string{exp, test.Exp}, "|")
			}
			if flags.String() != exp {
				t.Errorf("values don't match %d: (%o) %s, %s", c, flags, flags, exp)
			}

			f, err := ParseFlags(exp)
			if err != nil || f != flags {
				if err != nil {
					t.Errorf("parsing error %q", err)
				} else {
					t.Errorf("error parsing %d: (%o), %q, (%o), %q", c, f, f, flags, exp)
				}
			}
			c++
		}
	}
}

func combinationsOf(input []testType) []testType {
	listCh := make(chan []testType)
	go func() {
		defer close(listCh)
		for i := 0; i <= len(input); i++ {
			combine(listCh, input, i)
		}
	}()

	var out []testType
	for list := range listCh {
		var f Flags
		names := make([]string, len(list))
		for i, test := range list {
			f |= test.In
			names[i] = test.Exp
		}
		out = append(out, testType{f, strings.Join(names, "|")})
	}
	return out

}

// combine - recursively builds all `size` combinations of elements of `input`
func combine(out chan []testType, input []testType, size int, values ...testType) {

	if len(values) == size {
		out <- append([]testType{}, values...)
		return
	}

	for j, v := range input {
		combine(out, input[j+1:], size, append(values, v)...)
	}
}
