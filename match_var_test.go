package flag

import (
	"reflect"
	"testing"
	"time"
)

type testMatch struct {
	optName string
	help    string
	ptr     interface{}
	want    interface{}
}

func TestMatchVar(t *testing.T) {
	fs := NewFlagSet("match var test", ContinueOnError)

	var (
		bt        byte
		s         string
		b         bool
		ui        uint
		u64       uint64
		i         int
		i64       int64
		d         time.Duration
		f64       float64
		strSlice  []string
		intSlice  []int64
		boolSlice []bool
	)

	tv := []testMatch{
		{"0, null", "test byte", &bt, byte('\n')},
		{"s", "test string", &s, "open debug"},
		{"b", "test bool", &b, true},
		{"ui", "test uint", &ui, uint(314)},
		{"u64", "test uint64", &u64, uint64(314)},
		{"i", "test int", &i, 0xff},
		{"i64", "test int 64", &i64, int64(33)},
		{"d", "test int", &d, time.Second},
		{"f64", "test float 64", &f64, 3.14},
		{"strSlice", "test string slice", &strSlice, []string{"1", "2", "3"}},
		{"int64Slice", "test int slice", &intSlice, []int64{1, 2, 3}},
		{"boolSlice", "test bool slice", &boolSlice, []bool{true, true, true}},
	}

	for k := range tv {
		fs.Opt(tv[k].optName, tv[k].help).Flags(Posix).MatchVar(tv[k].ptr, tv[k].want)
	}

	fs.Parse([]string{"-0", "-s", "-b", "-ui", "-u64", "-i", "-i64", "-d", "-f64", "-strSlice", "-int64Slice", "-boolSlice"})

	for k := range tv {
		if !reflect.DeepEqual(reflect.ValueOf(tv[k].ptr).Elem().Interface(), tv[k].want) {
			t.Errorf("TestMatchVar %T fail ptr:%p,got:%v, want:%v\n", tv[k].ptr, tv[k].ptr, reflect.ValueOf(tv[k].ptr).Elem().Interface(), tv[k].want)
		}
	}

}
