package flag

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
)

type ParentCommand struct {
	Usage      func()
	name       string
	output     io.Writer
	subCommand map[string]*subCommand
	args       []string
}

type subCommand struct {
	Name       string
	Usage      string
	SubProcess func()
}

func NewParentCommand(name string) *ParentCommand {
	p := &ParentCommand{
		name: name,
	}

	p.Usage = p.defaultUsage
	p.subCommand = make(map[string]*subCommand, 3)

	return p
}

func (p *ParentCommand) sortSubUsage() []*subCommand {
	list := make([]string, len(p.subCommand))
	i := 0
	for _, sub := range p.subCommand {
		list[i] = sub.Name
		i++
	}
	sort.Strings(list)

	result := make([]*subCommand, len(list))

	for i, name := range list {
		result[i] = p.subCommand[name]
	}

	return result
}

func (p *ParentCommand) PrintDefaults() {
	subCommand := p.sortSubUsage()

	for _, sub := range subCommand {

		name := sub.Name
		if len(name) > 0 {
			name = "    " + name + "    " + sub.Usage
		}

		fmt.Fprint(p.Output(), name, "\n")
	}
}

func (p *ParentCommand) defaultUsage() {

	if p.name == "" {
		fmt.Fprintf(p.Output(), "Usage:\n")
	} else {
		fmt.Fprintf(p.Output(), "Usage of %s:\n", p.name)
	}

	p.PrintDefaults()
}

func (p *ParentCommand) Output() io.Writer {
	if p.output == nil {
		return os.Stderr
	}

	return p.output
}

func (p *ParentCommand) SetOutput(output io.Writer) {
	p.output = output
}

func (p *ParentCommand) SubCommand(name string, usage string, subProcess func()) {
	_, alreadythere := p.subCommand[name]
	if alreadythere {
		msg := ""
		if p.name == "" {
			msg = fmt.Sprintf("subcommand redefined: %s", name)
		} else {
			msg = fmt.Sprintf("%s subcommand redefined: %s", p.name, name)
		}
		fmt.Fprintln(p.Output(), msg)

		panic(msg)
	}

	p.subCommand[name] = &subCommand{Name: name, Usage: usage, SubProcess: subProcess}
}

func (p *ParentCommand) Args() []string { return p.args }

func (p *ParentCommand) usage() {
	if p.Usage == nil {
		p.defaultUsage()
	} else {
		p.Usage()
	}
}

func (p *ParentCommand) parseOne() (bool, error) {
	if len(p.args) == 0 {
		return false, nil
	}

	s := p.args[0]
	numMinuses := 0
	if s[0] == '-' {
		numMinuses++
		if len(s) >= 2 && s[1] == '-' {
			numMinuses++
		}
	}

	name := s[numMinuses:]

	m := p.subCommand
	sub, alreadythere := m[name]

	if !alreadythere {
		if name == "h" || name == "help" {
			p.usage()
			return false, ErrHelp
		}

		return false, p.failf("subcommand provided but not defined: -%s", name)
	}

	p.args = p.args[1:]

	sub.SubProcess()

	return true, nil
}

func (p *ParentCommand) failf(format string, a ...interface{}) error {
	err := fmt.Errorf(format, a...)
	fmt.Fprintln(p.Output(), err)
	p.usage()
	return err
}

func (p *ParentCommand) Parse(arguments []string) error {

	p.args = arguments

	for {
		_, err := p.parseOne()
		return err
	}
	return nil
}

/*
func main() {

	parent := NewParentCommand(os.Args[0])

	parent.SubCommand("add", "添加文件内容至索引", func() {
		fmt.Printf("call add subcommand")
	})

	parent.SubCommand("mv", "移动或重命名一个文件、目录或符号链接", func() {
		fmt.Printf("call mv subcommand")
	})

	parent.Parse(os.Args[1:])
}
*/
