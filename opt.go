package flag

import (
	"time"
)

func (f *FlagSet) flagVar(flag *Flag) {

	if flag.flags&PosixShort > 0 && flag.flags&GreedyMode > 0 {
		panic("Cannot set both PosixShort and GreedyMode")
	}

	name := flag.Name
	name, names, ok := newName(name)
	if ok {
		initFormal(&f.formal2)
		for _, v := range names {
			_, alreadythere := f.formal2[v]
			if alreadythere {
				f.alreadythereError(v)
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

func Opt(name string, usage string) {
	CommandLine.Opt(name, usage)
}
