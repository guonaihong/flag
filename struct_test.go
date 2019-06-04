package flag

import (
	"sort"
	"testing"
	"time"
)

type structOptiont1 struct {
	Y int `opt:"y,year" defValue:"2019" usage:"test y"`
}

type structOptiont2 struct {
	M int `opt:"M,month" usage:"test m"`
}

type structOption struct {
	Header       []string      `opt:"H, header" flags:"posix" usage:"http header"`
	MaxThreads   int           `opt:"m, max" flags:"posix" defValue:"1" usage:"max threads"`
	Packet       []string      `opt:"p, packet" flags:"greedy" usage:"websocket packet"`
	Debug        bool          `opt:"d, debug" defValue:"true" usage:"open debug mode"`
	Int          int           `opt:"i, int" defValue:"31" usage:"test int type"`
	Int64        int64         `opt:"i64, int64" defValue:"33" usage:"test int64 type"`
	Float64      float64       `opt:"f64, float64" defValue:"1.1" usage:"test floa64 type"`
	String       string        `opt:"s, string" defValue:"test string" usage:"test string type"`
	Uint         uint          `opt:"u, uint" defValue:"17" usage:"test uint type"`
	Uint64       uint64        `opt:"u64, uint64" defValue:"13" usage:"test uint64 type"`
	StringSlice  []string      `opt:"ss, string-slice" defValue:"aa,bb" usage:"string slice type"`
	StringSlice2 []string      `opt:"ss2, string-slice2" defValue:"aa,bb#cc,dd" sep:"#" usage:"string slice type"`
	Int64Slice   []int64       `opt:"i64s, int64-slice" defValue:"1,2" usage:"int slice type"`
	T            time.Duration `opt:"t, time" defValue:"1s" usage:"test time.Duration type"`
	structOptiont1
	structOptiont2
}

// Test Module-name func-name
func TestStructParse(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)

	o := structOption{}

	fs.ParseStruct([]string{"-H", "appkey:value1", "-H", "user:username", "-H", "passwd:passwd", "-m", "10", "-p", "@./test.data", "-M", "6"}, &o)

	if o.MaxThreads != 10 {
		t.Errorf("maxThreads got %d want 10, %p\n", o.MaxThreads, &o.MaxThreads)
	}

	if len(o.Packet) != 1 || o.Packet[0] != "@./test.data" {
		t.Errorf("Packet got %s want @./test.data\n", o.Packet)
	}

	if len(o.Header) != 3 {
		t.Errorf("Header got %s want appkey:value1, user:username passwd:passwd\n", o.Header)
	} else {
		need := []string{"appkey:value1", "user:username", "passwd:passwd"}
		sort.Strings(o.Header)
		sort.Strings(need)

		if o.Header[0] != need[0] {
			t.Errorf("Header got %s want %s\n", o.Header[0], need[0])
		}
		if o.Header[1] != need[1] {
			t.Errorf("Header got %s want %s\n", o.Header[1], need[1])
		}
		if o.Header[2] != need[2] {
			t.Errorf("Header got %s want %s\n", o.Header[2], need[2])
		}
	}

	if !o.Debug {
		t.Errorf("debug got %t want true\n", o.Debug)
	}

	if o.Int != 31 {
		t.Errorf("int got %d want 31\n", o.Int)
	}

	if o.Int64 != 33 {
		t.Errorf("Int64 got %d want 33\n", o.Int64)
	}

	if o.Float64 != 1.1 {
		t.Errorf("float64 got %f want 1.1\n", o.Float64)
	}

	if o.String != "test string" {
		t.Errorf("string got %s want test string\n", o.String)
	}

	if o.Uint != 17 {
		t.Errorf("uint got %d want 17\n", o.Uint)
	}

	if o.Uint64 != 13 {
		t.Errorf("uint64 got %d want 13\n", o.Uint64)
	}

	if len(o.StringSlice) != 2 {
		t.Errorf("len(%d) uint64 got %v want [\"aa\", \"bb\"]\n", len(o.StringSlice), o.StringSlice)
	} else {
		if o.StringSlice[0] != "aa" {
			t.Errorf("stringSlice got %s want aa\n", o.StringSlice[0])
		}

		if o.StringSlice[1] != "bb" {
			t.Errorf("stringSlice got %s want bb\n", o.StringSlice[1])
		}
	}

	if len(o.StringSlice2) != 2 {
		t.Errorf("len(%d) uint64 got %v want [\"aa,bb\", \"cc,dd\"]\n", len(o.StringSlice2), o.StringSlice2)
	} else {
		if o.StringSlice2[0] != "aa,bb" {
			t.Errorf("stringSlice got %s want aa,bb\n", o.StringSlice2[0])
		}

		if o.StringSlice2[1] != "cc,dd" {
			t.Errorf("stringSlice got %s want cc,dd\n", o.StringSlice2[1])
		}
	}

	if len(o.Int64Slice) != 2 {
		t.Errorf(`len(%d) uint64 got %v want ["1", "2"]`+"\n", len(o.Int64Slice), o.Int64Slice)
	} else {
		if o.Int64Slice[0] != 1 {
			t.Errorf("intSlice got %d want 1\n", o.Int64Slice[0])
		}

		if o.Int64Slice[1] != 2 {
			t.Errorf("intSlice got %d want 2\n", o.Int64Slice[1])
		}
	}

	if int64(o.T) != int64(time.Second) {
		t.Errorf("got %v want 1s\n", o.T)
	}

	if o.Y != 2019 {
		t.Errorf("got %v want 2019\n", o.Y)
	}

	if o.M != 6 {
		t.Errorf("got %v want 6\n", o.M)
	}
}
