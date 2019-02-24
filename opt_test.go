package flag

import (
	"testing"
)

func TestPosixShort(t *testing.T) {
	fs := NewFlagSet("cat", ContinueOnError)
	showNonprinting := fs.Opt("v, show-nonprinting", "use ^ and M- notation, except for LFD and TAB").
		Flags(PosixShort).NewBool(false)

	showTabs := fs.Opt("T, show-tabs", "display TAB characters as ^I").Flags(PosixShort).NewBool(false)
	fs.Parse([]string{"-Tv"})

	if *showNonprinting != true && *showTabs != true {
		t.Error("flag was not set by -Tv")
	}

	*showNonprinting = false
	*showTabs = false
	fs.Parse([]string{"-Tv", "--show-nonprinting", "--show-tabs"})
	if *showNonprinting != true && *showTabs != true {
		t.Error("flag was not set by --show-nonprinting --show-tabs")
	}

	opts := make([]*bool, 7)
	opts[0] = fs.Opt("A, show-all", "equivalent to -vET").Flags(PosixShort).NewBool(false)
	opts[1] = fs.Opt("b, number-nonblank", "number nonempty output lines, overrides -n").Flags(PosixShort).NewBool(false)
	opts[2] = fs.Opt("e", "equivalent to -vE").Flags(PosixShort).NewBool(false)
	opts[3] = fs.Opt("E, show-end", "display $ at end of each line").Flags(PosixShort).NewBool(false)
	opts[4] = fs.Opt("n, numbe", "number all output line").Flags(PosixShort).NewBool(false)
	opts[5] = fs.Opt("s, squeeze-blank", "suppress repeated empty output lines").Flags(PosixShort).NewBool(false)
	opts[6] = fs.Opt("t", "equivalent to -vT").Flags(PosixShort).NewBool(false)

	fs.Parse([]string{"-tsnEebA"})

	for _, v := range opts {
		if !*v {
			t.Error("flag was not set by -tsnEebA")
		}
	}

	err := fs.Parse([]string{"-m"})
	if err == nil {
		t.Fatal("error expected")
	}

	err = fs.Parse([]string{"-mA"})
	if err != nil {
		t.Fatal("error expected")
	}
}

func TestOptHelp(t *testing.T) {
	fs := NewFlagSet("cat", ContinueOnError)
	_ = fs.Opt("T, show-tabs", "display TAB characters as ^I").Flags(PosixShort).NewBool(false)
	err := fs.Parse([]string{"-vT, -help"})
	if err != nil {
		t.Fatal("expected no error; got ", err)
	}

	err = fs.Parse([]string{"-h"})
	if err != ErrHelp {
		t.Fatal("expected no error; got ", err)
	}
}
