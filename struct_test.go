package flag

import (
	"sort"
	"testing"
)

type structOption struct {
	Header     []string `opt:"h, header" flags:"posix" usage:"http header"`
	MaxThreads int      `opt:"m, max" flags:"posix" defValue:"1" usage:"max threads"`
	Packet     []string `opt:"p, packet" flags:"greedy" usage:"websocket packet"`
}

// Test Module-name func-name
func TestStructParse(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)

	o := structOption{}

	fs.ParseStruct([]string{"-h", "appkey:value1", "-h", "user:username", "-h", "passwd:passwd", "-m", "10", "-p", "@./test.data"}, &o)

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
}
