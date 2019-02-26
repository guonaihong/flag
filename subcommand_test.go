package flag

import (
	"testing"
)

func TestSubCommand(t *testing.T) {
	parent := NewParentCommand("test-subcommand")

	clone, init, add, mv, reset, rm := false, false, false, false, false, false
	parent.SubCommand("clone", "Clone a repository into a new directory", func() {
		t.Logf("call clone subcommand")
		clone = true
	})

	parent.SubCommand("init", "Create an empty Git repository or reinitialize an existing one", func() {
		t.Logf("call init subcommand")
		init = true
	})

	parent.SubCommand("add", "Add file contents to the index", func() {
		t.Logf("call add subcommand")
		add = true
	})

	parent.SubCommand("mv", "Move or rename a file, a directory, or a symlink", func() {
		t.Logf("call mv subcommand")
		mv = true
	})

	parent.SubCommand("reset", "Reset current HEAD to the specified state", func() {
		t.Logf("call reset subcommand")
		reset = true
	})

	parent.SubCommand("rm", "Remove files from the working tree and from the index", func() {
		t.Logf("call rm subcommand")
		rm = true
	})

	parent.Parse([]string{"clone"})
	parent.Parse([]string{"init"})
	parent.Parse([]string{"add"})
	parent.Parse([]string{"mv"})
	parent.Parse([]string{"reset"})
	parent.Parse([]string{"rm"})

	if !clone {
		t.Error("clone should be true")
	}

	if !init {
		t.Error("init should be true")
	}

	if !add {
		t.Error("add should be true")
	}
	if !mv {
		t.Error("mv should be true")
	}
	if !reset {
		t.Error("reset should be true")
	}
	if !rm {
		t.Error("rm should be true")
	}
}
