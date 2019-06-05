package flag

import (
	"encoding/json"
	"strconv"
)

type boolSlice []bool

func newBoolSliceValue(val []bool, p *[]bool) *boolSlice {
	*p = val
	return (*boolSlice)(p)
}

func (b *boolSlice) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}

	*b = append(*b, v)
	return nil
}

func (b *boolSlice) String() string {
	all, err := json.Marshal(b)
	if err != nil {
		panic(err.Error())
	}

	return string(all)
}

func (b *boolSlice) Get() interface{} {
	return []bool(*b)
}

type int64SliceValue []int64

func newInt64SliceValue(val []int64, p *[]int64) *int64SliceValue {
	*p = val
	return (*int64SliceValue)(p)
}

func (i *int64SliceValue) Set(val string) error {
	var iv int64Value

	err := iv.Set(val)
	if err != nil {
		return err
	}

	*i = append(*i, int64(iv))
	return nil
}

func (i *int64SliceValue) Get() interface{} {
	return []int64(*i)
}

func (i *int64SliceValue) String() string {
	all, err := json.Marshal(i)
	if err != nil {
		panic(err.Error())
	}
	return string(all)
}

// Int64SliceVar defines an int64 flag with specified name, default value, and usage string.
// The argument p points to an int64 slice variable in which to store the value of the flag.
func (f *FlagSet) Int64SliceVar(p *[]int64, name string, value []int64, usage string) {
	f.Var(newInt64SliceValue(value, p), name, usage)
}

// Int64SliceVar defines an int64 flag with specified name, default value, and usage string.
// The argument p points to an int64 slice variable in which to store the value of the flag.
func Int64SliceVar(p *[]int64, name string, value []int64, usage string) {
	CommandLine.Var(newInt64SliceValue(value, p), name, usage)
}

// Int64Slice defines an int64 flag with specified name, default value, and usage string.
// The return value is the address of an int64 slice variable that stores the value of the flag.
func (f *FlagSet) Int64Slice(name string, value []int64, usage string) *[]int64 {
	p := new([]int64)
	f.Int64SliceVar(p, name, value, usage)
	return p
}

// Int64Slice defines an int64 flag with specified name, default value, and usage string.
// The return value is the address of an int64 slice variable that stores the value of the flag.
func Int64Slice(name string, value []int64, usage string) *[]int64 {
	return CommandLine.Int64Slice(name, value, usage)
}

// -- string slice value
type stringSliceValue []string

func newStringSliceValue(val []string, p *[]string) *stringSliceValue {
	*p = val
	return (*stringSliceValue)(p)
}

func (s *stringSliceValue) Set(val string) error {
	*s = append(*s, val)
	return nil
}

func (s *stringSliceValue) Get() interface{} {
	return []string(*s)
}

func (s *stringSliceValue) String() string {
	all, err := json.Marshal(s)
	if err != nil {
		panic(err.Error())
	}
	return string(all)
}

// StringSliceVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string slice variable in which to store the value of the flag.
func (f *FlagSet) StringSliceVar(p *[]string, name string, value []string, usage string) {
	f.Var(newStringSliceValue(value, p), name, usage)
}

// StringSliceVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string slice variable in which to store the value of the flag.
func StringSliceVar(p *[]string, name string, value []string, usage string) {
	CommandLine.Var(newStringSliceValue(value, p), name, usage)
}

// StringSlice defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string slice variable that stores the value of the flag.
func (f *FlagSet) StringSlice(name string, value []string, usage string) *[]string {
	p := new([]string)
	f.StringSliceVar(p, name, value, usage)
	return p
}

// StringSlice defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string slice variable that stores the value of the flag.
func StringSlice(name string, value []string, usage string) *[]string {
	return CommandLine.StringSlice(name, value, usage)
}
