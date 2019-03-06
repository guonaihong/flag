package flag

import (
	"bytes"
	"time"
)

func (f *FlagSet) setNamesToMap(m *map[string]*Flag, names []string, flag *Flag) {

	initFormal(m)
	for _, v := range names {
		_, alreadythere := (*m)[v]
		if alreadythere {
			f.alreadythereError(v)
		}

		(*m)[v] = &Flag{Name: v,
			Usage:    flag.Usage,
			Value:    flag.Value,
			DefValue: flag.DefValue,
			flags:    flag.flags,
		}
	}
}

func (f *FlagSet) flagVar(flag *Flag) {

	if flag.flags&PosixShort > 0 && flag.flags&GreedyMode > 0 {
		panic("Cannot set both PosixShort and GreedyMode")
	}

	name := flag.Name
	var names []string
	var ok bool

	if flag.isOption {
		f.setNamesToMap(&f.regex, []string{flag.Regex}, flag)
		f.setNamesToMap(&f.shortLong, flag.Short, flag)
		f.setNamesToMap(&f.shortLong, flag.Long, flag)
		name = flag.Name
	} else {
		name, names, ok = newName(name)
		if ok {
			f.setNamesToMap(&f.shortLong, names, flag)
		}
	}

	_, alreadythere := f.formal[name]
	if alreadythere {
		f.alreadythereError(name)
	}

	initFormal(&f.formal)

	f.formal[name] = flag
}

func (f *FlagSet) OptOpt(opt Flag) *Flag {
	var buf bytes.Buffer

	if len(opt.Name) > 0 {
		panic("OptOpt function does not support setting Name")
	}

	if len(opt.Regex) > 0 {
		buf.WriteString(opt.Regex)
	}

	if len(opt.Short) > 0 {
		if buf.Len() > 0 {
			buf.WriteString(", ")
		}

		for k, v := range opt.Short {
			buf.WriteString(v)
			if len(opt.Short) != k+1 {
				buf.WriteString(", ")
			}
		}
	}

	if len(opt.Long) > 0 {
		if buf.Len() > 0 {
			buf.WriteString(", ")
		}

		for k, v := range opt.Long {
			buf.WriteString(v)
			if len(opt.Long) != k+1 {
				buf.WriteString(", ")
			}
		}
	}

	if buf.Len() > 0 {
		opt.isOption = true
		opt.Name = buf.String()
	}

	opt.parent = f
	return &opt
}

func (f *FlagSet) Opt(name string, usage string) *Flag {
	return &Flag{Name: name, Usage: usage, parent: f}
}

func (f *Flag) Flags(flag Flags) *Flag {
	f.flags |= flag
	if flag&PosixShort == PosixShort {
		f.parent.openPosixShort = true
	}
	return f
}

func (f *Flag) NewBool(defValue bool) *bool {
	p := new(bool)
	f.Value = newBoolValue(defValue, p)
	f.parent.flagVar(f)
	return p
}

func (f *Flag) NewString(defValue string) *string {
	p := new(string)
	f.Value = newStringValue(defValue, p)
	f.parent.flagVar(f)
	return p
}

func (f *Flag) NewUint(defValue uint) *uint {
	p := new(uint)
	f.Value = newUintValue(defValue, p)
	f.parent.flagVar(f)
	return p
}

func (f *Flag) NewUint64(defValue uint64) *uint64 {
	p := new(uint64)
	f.Value = newUint64Value(defValue, p)
	f.parent.flagVar(f)
	return p
}

func (f *Flag) NewInt(defValue int) *int {
	p := new(int)
	f.Value = newIntValue(defValue, p)
	f.parent.flagVar(f)
	return p
}

func (f *Flag) NewInt64(defValue int64) *int64 {
	p := new(int64)
	f.Value = newInt64Value(defValue, p)
	f.parent.flagVar(f)
	return p
}

func (f *Flag) NewFloat64(defValue float64) *float64 {
	p := new(float64)
	f.Value = newFloat64Value(defValue, p)
	f.parent.flagVar(f)
	return p
}

func (f *Flag) NewDuration(defValue time.Duration) *time.Duration {
	p := new(time.Duration)
	f.Value = newDurationValue(defValue, p)
	f.parent.flagVar(f)
	return p
}

func (f *Flag) NewInt64Slice(defValue []int64) *[]int64 {
	p := new([]int64)
	f.Value = newInt64SliceValue(defValue, p)
	f.parent.flagVar(f)
	return p
}

func (f *Flag) NewStringSlice(defValue []string) *[]string {
	p := new([]string)
	f.Value = newStringSliceValue(defValue, p)
	f.parent.flagVar(f)
	return p
}

func Opt(name string, usage string) *Flag {
	return CommandLine.Opt(name, usage)
}
