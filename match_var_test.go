package flag

import (
	"testing"
)

type testMatch struct {
	Delimiter byte
	Limit     int
}

func TestMatchVar(t *testing.T) {
	fs := NewFlagSet("match var test", ContinueOnError)

	t := testMatch{
		Delimiter: '\n',
		Limit:     100,
	}

	command.Opt("0, null", "end each output line with NUL, not newline").
		Flags(Posix).MatchVar(&t.Delimiter)

	command.Opt("limit", "test int").
		Flags(Posix).MatchVar(&t.Limit)

	fs.Parse([]string{"-0", "-limit"})
}
