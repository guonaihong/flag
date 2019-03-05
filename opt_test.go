package flag

import (
	"testing"
	"time"
)

func TestPosixShortString(t *testing.T) {

	var ignoreCase *bool
	var afterContext *string

	fs := NewFlagSet("grep", ContinueOnError)
	addOption := func(fs *FlagSet) {
		ignoreCase = fs.Opt("i, ignore-case", "Ignore case distinctions,"+
			"so that characters that differ only in case match each other.").
			Flags(PosixShort).NewBool(false)

		afterContext = fs.Opt("A, after-context", "Print NUM lines of trailing context after matching lines."+
			"Places a line containing a group separator (--) between contiguous groups of matches.  "+
			"With the  -o or --only-matching option, this has no effect and a warning is given.").
			Flags(PosixShort).NewString("")
	}

	addOption(fs)
	fs.Parse([]string{"-iA5"})

	if !*ignoreCase || *afterContext != "5" {
		t.Error("flag was not set by -iA5")
	}

	fs = NewFlagSet("grep", ContinueOnError)

	addOption(fs)
	fs.Parse([]string{"-iA", "66"})

	if !*ignoreCase || *afterContext != "66" {
		t.Error("flag was not set by -iA 66")
	}
}

func TestPosixShortBool(t *testing.T) {
	fs := NewFlagSet("cat", ContinueOnError)
	showNonprinting := fs.Opt("v, show-nonprinting", "use ^ and M- notation, except for LFD and TAB").
		Flags(PosixShort).NewBool(false)

	showTabs := fs.Opt("T, show-tabs", "display TAB characters as ^I").Flags(PosixShort).NewBool(false)
	fs.Parse([]string{"-Tv"})

	if *showNonprinting != true || *showTabs != true {
		t.Error("flag was not set by -Tv")
	}

	*showNonprinting = false
	*showTabs = false
	fs.Parse([]string{"-Tv", "--show-nonprinting", "--show-tabs"})
	if *showNonprinting != true && *showTabs != true {
		t.Error("flag was not set by --show-nonprinting --show-tabs")
	}

	opts := make([]*bool, 7)
	opts[0] = fs.Opt("A, show-all", "equivalent to -vET").
		Flags(PosixShort).NewBool(false)
	opts[1] = fs.Opt("b, number-nonblank", "number nonempty output lines, overrides -n").
		Flags(PosixShort).NewBool(false)
	opts[2] = fs.Opt("e", "equivalent to -vE").
		Flags(PosixShort).NewBool(false)
	opts[3] = fs.Opt("E, show-end", "display $ at end of each line").
		Flags(PosixShort).NewBool(false)
	opts[4] = fs.Opt("n, numbe", "number all output line").
		Flags(PosixShort).NewBool(false)
	opts[5] = fs.Opt("s, squeeze-blank", "suppress repeated empty output lines").
		Flags(PosixShort).NewBool(false)
	opts[6] = fs.Opt("t", "equivalent to -vT").
		Flags(PosixShort).NewBool(false)

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

func TestOptGreedyMode(t *testing.T) {
	fs := NewFlagSet("test-custonslice", ContinueOnError)

	header := fs.Opt("H", "http header").Flags(GreedyMode).
		NewStringSlice([]string{})
	url := fs.Opt("url", "http url").NewString("")

	fs.Parse([]string{"-H", "sid:sid1234", "time:time-value", "score:1.0", "-url", "test.com"})

	testHeader := []string{"sid:sid1234", "time:time-value", "score:1.0"}

	if len(*header) != len(testHeader) {
		t.Fatal("The parsed header is inconsistent with testHeader",
			header, "\n",
			testHeader, "\n")
	}

	for k := range *header {
		if (*header)[k] != testHeader[k] {
			t.Fatal("The parsed header is inconsistent with testHeader",
				header, "\n",
				testHeader, "\n")
		}
	}

	if *url != "test.com" {
		t.Fatal("url fail->", *url, "\n")
	}

	args := fs.Args()
	if len(args) != 0 {
		t.Fatalf("len(args) != 0\n")
	}

	//	fs2 := NewFlagSet("jm").
}

func TestOptParse(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)

	boolFlag := fs.Opt("bool", "bool value").NewBool(false)
	bool2Flag := fs.Opt("bool2", "bool2 value").NewBool(false)
	intFlag := fs.Opt("int", "int value").NewInt(0)
	int64Flag := fs.Opt("int64", "int64 value").NewInt64(0)
	uintFlag := fs.Opt("uint", "uint value").NewUint(0)
	uint64Flag := fs.Opt("uint64", "uint64 value").NewUint64(0)
	stringFlag := fs.Opt("string", "string value").NewString("0")
	float64Flag := fs.Opt("float64", "float64 value").NewFloat64(0)
	durationFlag := fs.Opt("duration", "time.Duration value").NewDuration(5 * time.Second)

	extra := "one-extra-argument"
	args := []string{
		"-bool",
		"-bool2=true",
		"--int", "22",
		"--int64", "0x23",
		"-uint", "24",
		"--uint64", "25",
		"-string", "hello",
		"-float64", "2718e28",
		"-duration", "2m",
		extra,
	}

	if err := fs.Parse(args); err != nil {
		t.Fatal(err)
	}
	if !fs.Parsed() {
		t.Error("f.Parse() = false after Parse")
	}
	if *boolFlag != true {
		t.Error("bool flag should be true, is ", *boolFlag)
	}
	if *bool2Flag != true {
		t.Error("bool2 flag should be true, is ", *bool2Flag)
	}
	if *intFlag != 22 {
		t.Error("int flag should be 22, is ", *intFlag)
	}
	if *int64Flag != 0x23 {
		t.Error("int64 flag should be 0x23, is ", *int64Flag)
	}
	if *uintFlag != 24 {
		t.Error("uint flag should be 24, is ", *uintFlag)
	}
	if *uint64Flag != 25 {
		t.Error("uint64 flag should be 25, is ", *uint64Flag)
	}
	if *stringFlag != "hello" {
		t.Error("string flag should be `hello`, is ", *stringFlag)
	}
	if *float64Flag != 2718e28 {
		t.Error("float64 flag should be 2718e28, is ", *float64Flag)
	}
	if *durationFlag != 2*time.Minute {
		t.Error("duration flag should be 2m, is ", *durationFlag)
	}

	if len(fs.Args()) != 1 {
		t.Error("expected one argument, got", len(fs.Args()))
	} else if fs.Args()[0] != extra {
		t.Errorf("expected argument %q got %q", extra, fs.Args()[0])
	}

}
