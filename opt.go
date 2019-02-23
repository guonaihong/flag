package flag

import (
	"fmt"
)

func (f *FlagSet) flagVar(flag *Flag) {
	if flag.flags&PosixShort > 0 {
		if _, ok := flag.Value.(boolFlag); !ok {
			panic(fmt.Sprintf("Type %T cannot use PosixShort tag", flag.Value))
		}
	}

	name := flag.Name
	name, names, ok := newName(name)
	if ok {
		initFormal(&f.formal2)
		for _, v := range names {
			_, alreadythere := f.formal2[v]
			if alreadythere {
				f.alreadythereError(name)
			}

			f.formal2[v] = &Flag{Name: v,
				Usage:    flag.Usage,
				Value:    flag.Value,
				DefValue: flag.DefValue,
				cbs:      flag.cbs,
				flags:    flag.flags,
			}
		}
	}
	_, alreadythere := f.formal[name]
	if alreadythere {
		f.alreadythereError(name)
	}

	initFormal(&f.formal)

	f.formal[name] = flag
}

func (f *FlagSet) Opt(name string, usage string) *Flag {
	return &Flag{Name: name, Usage: usage, parent: f}
}

func (f *Flag) IsEnd(cb func()) *Flag {
	f.cbs = append(f.cbs, cb)
	return f
}

func (f *Flag) Flags(flag Flags) *Flag {
	f.flags |= flag
	f.parent.openPosixShort = true
	return f
}

func (f *Flag) NewBool(defValue bool) *bool {
	p := new(bool)
	f.Value = newBoolValue(defValue, p)
	f.parent.flagVar(f)
	return p
}

func (f *Flag) NewInt64Slice(defValue []int64) *[]int64 {
	p := new([]int64)
	f.Value = newInt64SliceValue(defValue, p)
	f.parent.flagVar(f)
	return p
}

func (f *Flag) NewStringlice(defValue []string) *[]string {
	p := new([]string)
	f.Value = newStringSliceValue(defValue, p)
	f.parent.flagVar(f)
	return p
}

func Opt(name string, usage string) {
	CommandLine.Opt(name, usage)
}
