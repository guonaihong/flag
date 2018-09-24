package flag

import (
	"os"
	"testing"
)

func TestSubCommand(t *testing.T) {
	parent := NewParentCommand(os.Args[0])

	parent.SubCommand("clone", "Clone a repository into a new directory", func() {
		t.Logf("call add subcommand")
	})

	parent.SubCommand("init", "Create an empty Git repository or reinitialize an existing one", func() {
		t.Logf("call add subcommand")
	})

	parent.SubCommand("add", "Add file contents to the index", func() {
		t.Logf("call add subcommand")
	})

	parent.SubCommand("mv", "Move or rename a file, a directory, or a symlink", func() {
		t.Logf("call mv subcommand")
	})

	parent.SubCommand("reset", "Reset current HEAD to the specified state", func() {
		t.Logf("call reset subcommand")
	})

	parent.SubCommand("rm", "Remove files from the working tree and from the index", func() {
		t.Logf("call reset subcommand")
	})
	parent.Parse(os.Args[1:])
}
