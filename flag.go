// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
	Package flag implements command-line flag parsing.

	Usage:

	Define flags using flag.String(), Bool(), Int(), etc.

	This declares an integer flag, -flagname, stored in the pointer ip, with type *int.
		import "flag"
		var ip = flag.Int("flagname", 1234, "help message for flagname")
	If you like, you can bind the flag to a variable using the Var() functions.
		var flagvar int
		func init() {
			flag.IntVar(&flagvar, "flagname", 1234, "help message for flagname")
		}
	Or you can create custom flags that satisfy the Value interface (with
	pointer receivers) and couple them to flag parsing by
		flag.Var(&flagVal, "name", "help message for flagname")
	For such flags, the default value is just the initial value of the variable.

	After all flags are defined, call
		flag.Parse()
	to parse the command line into the defined flags.

	Flags may then be used directly. If you're using the flags themselves,
	they are all pointers; if you bind to variables, they're values.
		fmt.Println("ip has value ", *ip)
		fmt.Println("flagvar has value ", flagvar)

	After parsing, the arguments following the flags are available as the
	slice flag.Args() or individually as flag.Arg(i).
	The arguments are indexed from 0 through flag.NArg()-1.

	Command line flag syntax:
		-flag
		-flag=x
		-flag x  // non-boolean flags only
	One or two minus signs may be used; they are equivalent.
	The last form is not permitted for boolean flags because the
	meaning of the command
		cmd -x *
	where * is a Unix shell wildcard, will change if there is a file
	called 0, false, etc. You must use the -flag=false form to turn
	off a boolean flag.

	Flag parsing stops just before the first non-flag argument
	("-" is a non-flag argument) or after the terminator "--".

	Integer flags accept 1234, 0664, 0x1234 and may be negative.
	Boolean flags may be:
		1, 0, t, f, T, F, true, false, TRUE, FALSE, True, False
	Duration flags accept any input valid for time.ParseDuration.

	The default set of command-line flags is controlled by
	top-level functions.  The FlagSet type allows one to define
	independent sets of flags, such as to implement subcommands
	in a command-line interface. The methods of FlagSet are
	analogous to the top-level functions for the command-line
	flag set.
*/
package flag

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ErrHelp is the error returned if the -help or -h flag is invoked
// but no such flag is defined.
var ErrHelp = errors.New("flag: help requested")
var ErrVersion = errors.New("flag: version requested")

// Value is the interface to the dynamic value stored in a flag.
// (The default value is represented as a string.)
//
// If a Value has an IsBoolFlag() bool method returning true,
// the command-line parser makes -name equivalent to -name=true
// rather than using the next command-line argument.
//
// Set is called once, in command line order, for each flag present.
// The flag package may call the String method with a zero-valued receiver,
// such as a nil pointer.
type Value interface {
	String() string
	Set(string) error
}

// Getter is an interface that allows the contents of a Value to be retrieved.
// It wraps the Value interface, rather than being part of it, because it
// appeared after Go 1 and its compatibility rules. All Value types provided
// by this package satisfy the Getter interface.
type Getter interface {
	Value
	Get() interface{}
}

type Flags int

const (
	PosixShort Flags = 1 << iota
	GreedyMode
	RegexKeyIsValue
	NotValue
)

// alias
const (
	Posix  = PosixShort
	Greedy = GreedyMode
)

// ErrorHandling defines how FlagSet.Parse behaves if the parse fails.
type ErrorHandling int

// These constants cause FlagSet.Parse to behave as described if the parse fails.
const (
	ContinueOnError ErrorHandling = iota // Return a descriptive error.
	ExitOnError                          // Call os.Exit(2).
	PanicOnError                         // Call panic with a descriptive error.
)

// A FlagSet represents a set of defined flags. The zero value of a FlagSet
// has no name and has ContinueOnError error handling.
type FlagSet struct {
	// Usage is the function called when an error occurs while parsing flags.
	// The field is a function (not a method) that may be changed to point to
	// a custom error handler. What happens after Usage is called depends
	// on the ErrorHandling setting; for the command line, this defaults
	// to ExitOnError, which exits the program after calling Usage.
	Usage func()

	name      string
	version   string
	author    string
	parsed    bool
	actual    map[string]*Flag
	formal    map[string]*Flag
	shortLong map[string]*Flag
	regex     map[string]*Flag

	args           []string // arguments after flags
	unkownArgs     []string // Unresolvable parameters
	errorHandling  ErrorHandling
	output         io.Writer // nil means stderr; use out() accessor
	openPosixShort bool
}

// A Flag represents the state of a flag.
type Flag struct {
	Name     string // name as it appears on command line
	Usage    string // help message
	Value    Value  // value as set
	DefValue string // default value (as text); for usage message

	pointer interface{}
	// 如果命令行匹配到Name，并且设置NotValue值，则使用MatchValue里面的值
	matchValue interface{} // (flags & NotValue) == NotValue

	parent *FlagSet
	flags  Flags

	Regex    string
	Short    []string
	Long     []string
	isOptOpt bool
}

// sortFlags returns the flags as a slice in lexicographical sorted order.
func sortFlags(flags map[string]*Flag) []*Flag {
	list := make(sort.StringSlice, len(flags))
	i := 0
	for _, f := range flags {
		list[i] = f.Name
		i++
	}
	list.Sort()
	result := make([]*Flag, len(list))
	for i, name := range list {
		result[i] = flags[name]
	}
	return result
}

// Output returns the destination for usage and error messages. os.Stderr is returned if
// output was not set or was set to nil.
func (f *FlagSet) Output() io.Writer {
	if f.output == nil {
		return os.Stderr
	}
	return f.output
}

// Name returns the name of the flag set.
func (f *FlagSet) Name() string {
	return f.name
}

// ErrorHandling returns the error handling behavior of the flag set.
func (f *FlagSet) ErrorHandling() ErrorHandling {
	return f.errorHandling
}

// SetOutput sets the destination for usage and error messages.
// If output is nil, os.Stderr is used.
func (f *FlagSet) SetOutput(output io.Writer) {
	f.output = output
}

// VisitAll visits the flags in lexicographical order, calling fn for each.
// It visits all flags, even those not set.
func (f *FlagSet) VisitAll(fn func(*Flag)) {
	for _, flag := range sortFlags(f.formal) {
		fn(flag)
	}
}

// VisitAll visits the command-line flags in lexicographical order, calling
// fn for each. It visits all flags, even those not set.
func VisitAll(fn func(*Flag)) {
	CommandLine.VisitAll(fn)
}

// Visit visits the flags in lexicographical order, calling fn for each.
// It visits only those flags that have been set.
func (f *FlagSet) Visit(fn func(*Flag)) {
	for _, flag := range sortFlags(f.actual) {
		fn(flag)
	}
}

// Visit visits the command-line flags in lexicographical order, calling fn
// for each. It visits only those flags that have been set.
func Visit(fn func(*Flag)) {
	CommandLine.Visit(fn)
}

// Lookup returns the Flag structure of the named flag, returning nil if none exists.
func (f *FlagSet) Lookup(name string) *Flag {
	return f.formal[name]
}

// Lookup returns the Flag structure of the named command-line flag,
// returning nil if none exists.
func Lookup(name string) *Flag {
	return CommandLine.formal[name]
}

// Set sets the value of the named flag.
func (f *FlagSet) Set(name, value string) error {
	flag, ok := f.formal[name]
	if !ok {
		return fmt.Errorf("no such flag -%v", name)
	}
	err := flag.Value.Set(value)
	if err != nil {
		return err
	}
	if f.actual == nil {
		f.actual = make(map[string]*Flag)
	}
	f.actual[name] = flag
	return nil
}

// Set sets the value of the named command-line flag.
func Set(name, value string) error {
	return CommandLine.Set(name, value)
}

// isZeroValue guesses whether the string represents the zero
// value for a flag. It is not accurate but in practice works OK.
func isZeroValue(flag *Flag, value string) bool {
	// Build a zero value of the flag's Value type, and see if the
	// result of calling its String method equals the value passed in.
	// This works unless the Value type is itself an interface type.
	typ := reflect.TypeOf(flag.Value)
	var z reflect.Value
	if typ.Kind() == reflect.Ptr {
		z = reflect.New(typ.Elem())
	} else {
		z = reflect.Zero(typ)
	}
	if value == z.Interface().(Value).String() {
		return true
	}

	switch value {
	case "false", "", "0":
		return true
	}
	return false
}

// UnquoteUsage extracts a back-quoted name from the usage
// string for a flag and returns it and the un-quoted usage.
// Given "a `name` to show" it returns ("name", "a name to show").
// If there are no back quotes, the name is an educated guess of the
// type of the flag's value, or the empty string if the flag is boolean.
func UnquoteUsage(flag *Flag) (name string, usage string) {
	// Look for a back-quoted name, but avoid the strings package.
	usage = flag.Usage
	for i := 0; i < len(usage); i++ {
		if usage[i] == '`' {
			for j := i + 1; j < len(usage); j++ {
				if usage[j] == '`' {
					name = usage[i+1 : j]
					usage = usage[:i] + name + usage[j+1:]
					return name, usage
				}
			}
			break // Only one back quote; use type name.
		}
	}
	// No explicit name, so use type if we can find one.
	name = "value"
	switch flag.Value.(type) {
	case boolFlag:
		name = ""
	case *durationValue:
		name = "duration"
	case *durationSliceValue:
		name = "duration[]"
	case *float64Value:
		name = "float"
	case *intValue, *int64Value:
		name = "int"
	case *stringValue:
		name = "string"
	case *stringSliceValue:
		name = "string[]"
	case *uintValue, *uint64Value:
		name = "uint"
	}
	return
}

// PrintDefaults prints, to standard error unless configured otherwise, the
// default values of all defined command-line flags in the set. See the
// documentation for the global function PrintDefaults for more information.
func (f *FlagSet) PrintDefaults() {
	f.VisitAll(func(flag *Flag) {
		name := strings.Replace(flag.Name, ", ", ", --", -1)
		s := fmt.Sprintf("  -%s", name) // Two spaces before -; see next two comments.
		name, usage := UnquoteUsage(flag)
		if len(name) > 0 {
			s += " " + name
		}
		// Boolean flags of one ASCII letter are so common we
		// treat them specially, putting their usage on the same line.
		if len(s) <= 4 { // space, space, '-', 'x'.
			s += "\t"
		} else {
			// Four spaces before the tab triggers good alignment
			// for both 4- and 8-space tab stops.
			s += "\n    \t"
		}
		s += strings.Replace(usage, "\n", "\n    \t", -1)

		if !isZeroValue(flag, flag.DefValue) {
			if _, ok := flag.Value.(*stringValue); ok {
				// put quotes on the value
				s += fmt.Sprintf(" (default %q)", flag.DefValue)
			} else {
				s += fmt.Sprintf(" (default %v)", flag.DefValue)
			}
		}
		fmt.Fprint(f.Output(), s, "\n")
	})
}

// PrintDefaults prints, to standard error unless configured otherwise,
// a usage message showing the default settings of all defined
// command-line flags.
// For an integer valued flag x, the default output has the form
//	-x int
//		usage-message-for-x (default 7)
// The usage message will appear on a separate line for anything but
// a bool flag with a one-byte name. For bool flags, the type is
// omitted and if the flag name is one byte the usage message appears
// on the same line. The parenthetical default is omitted if the
// default is the zero value for the type. The listed type, here int,
// can be changed by placing a back-quoted name in the flag's usage
// string; the first such item in the message is taken to be a parameter
// name to show in the message and the back quotes are stripped from
// the message when displayed. For instance, given
//	flag.String("I", "", "search `directory` for include files")
// the output will be
//	-I directory
//		search directory for include files.
func PrintDefaults() {
	CommandLine.PrintDefaults()
}

func (f *FlagSet) printVersion() {
	name := f.name
	if name != "" {
		name += " "
	}

	fmt.Fprintf(f.Output(), "%s%s\n", name, f.version)

}

func (f *FlagSet) printVersionAuthor() {
	if f.author != "" {
		fmt.Fprintf(f.Output(), "%s\n\n", f.author)
	}
}

// defaultUsage is the default function to print a usage message.
func (f *FlagSet) defaultUsage() {

	f.printVersionAuthor()

	if f.name == "" {
		fmt.Fprintf(f.Output(), "Usage:\n")
	} else {
		fmt.Fprintf(f.Output(), "Usage of %s:\n", f.name)
	}
	f.PrintDefaults()
}

// NOTE: Usage is not just defaultUsage(CommandLine)
// because it serves (via godoc flag Usage) as the example
// for how to write your own usage function.

// Usage prints a usage message documenting all defined command-line flags
// to CommandLine's output, which by default is os.Stderr.
// It is called when an error occurs while parsing flags.
// The function is a variable that may be changed to point to a custom function.
// By default it prints a simple header and calls PrintDefaults; for details about the
// format of the output and how to control it, see the documentation for PrintDefaults.
// Custom usage functions may choose to exit the program; by default exiting
// happens anyway as the command line's error handling strategy is set to
// ExitOnError.
var Usage = func() {
	CommandLine.printVersionAuthor()
	fmt.Fprintf(CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	PrintDefaults()
}

// NFlag returns the number of flags that have been set.
func (f *FlagSet) NFlag() int { return len(f.actual) }

// NFlag returns the number of command-line flags that have been set.
func NFlag() int { return len(CommandLine.actual) }

// Arg returns the i'th argument. Arg(0) is the first remaining argument
// after flags have been processed. Arg returns an empty string if the
// requested element does not exist.
func (f *FlagSet) Arg(i int) string {
	if i < 0 || i >= len(f.args) {
		return ""
	}
	return f.args[i]
}

// Arg returns the i'th command-line argument. Arg(0) is the first remaining argument
// after flags have been processed. Arg returns an empty string if the
// requested element does not exist.
func Arg(i int) string {
	return CommandLine.Arg(i)
}

// NArg is the number of arguments remaining after flags have been processed.
func (f *FlagSet) NArg() int { return len(f.args) }

// NArg is the number of arguments remaining after flags have been processed.
func NArg() int { return len(CommandLine.args) }

// Args returns the non-flag arguments.
func (f *FlagSet) Args() []string { return f.args }

// Args returns the non-flag command-line arguments.
func Args() []string { return CommandLine.args }

// BoolVar defines a bool flag with specified name, default value, and usage string.
// The argument p points to a bool variable in which to store the value of the flag.
func (f *FlagSet) BoolVar(p *bool, name string, value bool, usage string) {
	f.Var(newBoolValue(value, p), name, usage)
}

// BoolVar defines a bool flag with specified name, default value, and usage string.
// The argument p points to a bool variable in which to store the value of the flag.
func BoolVar(p *bool, name string, value bool, usage string) {
	CommandLine.Var(newBoolValue(value, p), name, usage)
}

// Bool defines a bool flag with specified name, default value, and usage string.
// The return value is the address of a bool variable that stores the value of the flag.
func (f *FlagSet) Bool(name string, value bool, usage string) *bool {
	p := new(bool)
	f.BoolVar(p, name, value, usage)
	return p
}

// Bool defines a bool flag with specified name, default value, and usage string.
// The return value is the address of a bool variable that stores the value of the flag.
func Bool(name string, value bool, usage string) *bool {
	return CommandLine.Bool(name, value, usage)
}

// IntVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func (f *FlagSet) IntVar(p *int, name string, value int, usage string) {
	f.Var(newIntValue(value, p), name, usage)
}

// IntVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func IntVar(p *int, name string, value int, usage string) {
	CommandLine.Var(newIntValue(value, p), name, usage)
}

// Int defines an int flag with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func (f *FlagSet) Int(name string, value int, usage string) *int {
	p := new(int)
	f.IntVar(p, name, value, usage)
	return p
}

// Int defines an int flag with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func Int(name string, value int, usage string) *int {
	return CommandLine.Int(name, value, usage)
}

// Int64Var defines an int64 flag with specified name, default value, and usage string.
// The argument p points to an int64 variable in which to store the value of the flag.
func (f *FlagSet) Int64Var(p *int64, name string, value int64, usage string) {
	f.Var(newInt64Value(value, p), name, usage)
}

// Int64Var defines an int64 flag with specified name, default value, and usage string.
// The argument p points to an int64 variable in which to store the value of the flag.
func Int64Var(p *int64, name string, value int64, usage string) {
	CommandLine.Var(newInt64Value(value, p), name, usage)
}

// Int64 defines an int64 flag with specified name, default value, and usage string.
// The return value is the address of an int64 variable that stores the value of the flag.
func (f *FlagSet) Int64(name string, value int64, usage string) *int64 {
	p := new(int64)
	f.Int64Var(p, name, value, usage)
	return p
}

// Int64 defines an int64 flag with specified name, default value, and usage string.
// The return value is the address of an int64 variable that stores the value of the flag.
func Int64(name string, value int64, usage string) *int64 {
	return CommandLine.Int64(name, value, usage)
}

// UintVar defines a uint flag with specified name, default value, and usage string.
// The argument p points to a uint variable in which to store the value of the flag.
func (f *FlagSet) UintVar(p *uint, name string, value uint, usage string) {
	f.Var(newUintValue(value, p), name, usage)
}

// UintVar defines a uint flag with specified name, default value, and usage string.
// The argument p points to a uint variable in which to store the value of the flag.
func UintVar(p *uint, name string, value uint, usage string) {
	CommandLine.Var(newUintValue(value, p), name, usage)
}

// Uint defines a uint flag with specified name, default value, and usage string.
// The return value is the address of a uint variable that stores the value of the flag.
func (f *FlagSet) Uint(name string, value uint, usage string) *uint {
	p := new(uint)
	f.UintVar(p, name, value, usage)
	return p
}

// Uint defines a uint flag with specified name, default value, and usage string.
// The return value is the address of a uint variable that stores the value of the flag.
func Uint(name string, value uint, usage string) *uint {
	return CommandLine.Uint(name, value, usage)
}

// Uint64Var defines a uint64 flag with specified name, default value, and usage string.
// The argument p points to a uint64 variable in which to store the value of the flag.
func (f *FlagSet) Uint64Var(p *uint64, name string, value uint64, usage string) {
	f.Var(newUint64Value(value, p), name, usage)
}

// Uint64Var defines a uint64 flag with specified name, default value, and usage string.
// The argument p points to a uint64 variable in which to store the value of the flag.
func Uint64Var(p *uint64, name string, value uint64, usage string) {
	CommandLine.Var(newUint64Value(value, p), name, usage)
}

// Uint64 defines a uint64 flag with specified name, default value, and usage string.
// The return value is the address of a uint64 variable that stores the value of the flag.
func (f *FlagSet) Uint64(name string, value uint64, usage string) *uint64 {
	p := new(uint64)
	f.Uint64Var(p, name, value, usage)
	return p
}

// Uint64 defines a uint64 flag with specified name, default value, and usage string.
// The return value is the address of a uint64 variable that stores the value of the flag.
func Uint64(name string, value uint64, usage string) *uint64 {
	return CommandLine.Uint64(name, value, usage)
}

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func (f *FlagSet) StringVar(p *string, name string, value string, usage string) {
	f.Var(newStringValue(value, p), name, usage)
}

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func StringVar(p *string, name string, value string, usage string) {
	CommandLine.Var(newStringValue(value, p), name, usage)
}

// String defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func (f *FlagSet) String(name string, value string, usage string) *string {
	p := new(string)
	f.StringVar(p, name, value, usage)
	return p
}

// String defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func String(name string, value string, usage string) *string {
	return CommandLine.String(name, value, usage)
}

// Float64Var defines a float64 flag with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the flag.
func (f *FlagSet) Float64Var(p *float64, name string, value float64, usage string) {
	f.Var(newFloat64Value(value, p), name, usage)
}

// Float64Var defines a float64 flag with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the flag.
func Float64Var(p *float64, name string, value float64, usage string) {
	CommandLine.Var(newFloat64Value(value, p), name, usage)
}

// Float64 defines a float64 flag with specified name, default value, and usage string.
// The return value is the address of a float64 variable that stores the value of the flag.
func (f *FlagSet) Float64(name string, value float64, usage string) *float64 {
	p := new(float64)
	f.Float64Var(p, name, value, usage)
	return p
}

// Float64 defines a float64 flag with specified name, default value, and usage string.
// The return value is the address of a float64 variable that stores the value of the flag.
func Float64(name string, value float64, usage string) *float64 {
	return CommandLine.Float64(name, value, usage)
}

// DurationVar defines a time.Duration flag with specified name, default value, and usage string.
// The argument p points to a time.Duration variable in which to store the value of the flag.
// The flag accepts a value acceptable to time.ParseDuration.
func (f *FlagSet) DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
	f.Var(newDurationValue(value, p), name, usage)
}

// DurationVar defines a time.Duration flag with specified name, default value, and usage string.
// The argument p points to a time.Duration variable in which to store the value of the flag.
// The flag accepts a value acceptable to time.ParseDuration.
func DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
	CommandLine.Var(newDurationValue(value, p), name, usage)
}

// Duration defines a time.Duration flag with specified name, default value, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the flag.
// The flag accepts a value acceptable to time.ParseDuration.
func (f *FlagSet) Duration(name string, value time.Duration, usage string) *time.Duration {
	p := new(time.Duration)
	f.DurationVar(p, name, value, usage)
	return p
}

// Duration defines a time.Duration flag with specified name, default value, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the flag.
// The flag accepts a value acceptable to time.ParseDuration.
func Duration(name string, value time.Duration, usage string) *time.Duration {
	return CommandLine.Duration(name, value, usage)
}

// DurationSliceVar defines a time.Duration flag with specified name, default value, and usage string.
// The argument p points to a time.Duration slice variable in which to store the value of the flag.
// The flag accepts a value acceptable to time.ParseDuration.
func (f *FlagSet) DurationSliceVar(p *[]time.Duration, name string, value []time.Duration, usage string) {
	f.Var(newDurationSliceValue(value, p), name, usage)
}

// DurationSliceVar defines a time.Duration flag with specified name, default value, and usage string.
// The argument p points to a time.Duration slice variable in which to store the value of the flag.
// The flag accepts a value acceptable to time.ParseDuration.
func DurationSliceVar(p *[]time.Duration, name string, value []time.Duration, usage string) {
	CommandLine.Var(newDurationSliceValue(value, p), name, usage)
}

// DurationSlice defines a time.Duration flag with specified name, default value, and usage string.
// The return value is the address of a time.Duration slice variable that stores the value of the flag.
// The flag accepts a value acceptable to time.ParseDuration.
func (f *FlagSet) DurationSlice(name string, value []time.Duration, usage string) *[]time.Duration {
	p := new([]time.Duration)
	f.DurationSliceVar(p, name, value, usage)
	return p
}

// DurationSlice defines a time.Duration flag with specified name, default value, and usage string.
// The return value is the address of a time.Duration slice variable that stores the value of the flag.
// The flag accepts a value acceptable to time.ParseDuration.
func DurationSlice(name string, value []time.Duration, usage string) *[]time.Duration {
	return CommandLine.DurationSlice(name, value, usage)
}

func newName(name string) (string, []string, bool) {
	pos := strings.Index(name, ",")
	if pos == -1 {
		return name, nil, false
	}

	names := strings.Split(name, ",")
	for k, _ := range names {
		names[k] = strings.TrimSpace(names[k])
	}

	sort.Slice(names, func(i, j int) bool {
		return len(names[i]) < len(names[j])
	})

	return strings.Join(names, ", "), names, true
}

func initFormal(formal *map[string]*Flag) {
	if *formal == nil {
		*formal = make(map[string]*Flag)
	}
}

func (f *FlagSet) alreadythereError(name string) {
	switch name {
	case "h", "help", "V", "version":
		return
	}
	var msg string
	if f.name == "" {
		msg = fmt.Sprintf("flag redefined: %s", name)
	} else {
		msg = fmt.Sprintf("%s flag redefined: %s", f.name, name)
	}
	fmt.Fprintln(f.Output(), msg)
	panic(msg) // Happens only if flags are declared with identical names
}

// Var defines a flag with the specified name and usage string. The type and
// value of the flag are represented by the first argument, of type Value, which
// typically holds a user-defined implementation of Value. For instance, the
// caller could create a flag that turns a comma-separated string into a slice
// of strings by giving the slice the methods of Value; in particular, Set would
// decompose the comma-separated string into the slice.
func (f *FlagSet) Var(value Value, name string, usage string) {
	// Remember the default value as a string; it won't change.
	name, names, ok := newName(name)
	flag := &Flag{Name: name, Usage: usage, Value: value, DefValue: value.String()}
	if ok {
		initFormal(&f.shortLong)
		for _, v := range names {
			_, alreadythere := f.shortLong[v]
			if alreadythere {
				f.alreadythereError(v)
			}

			f.shortLong[v] = &Flag{Name: v,
				Usage:    usage,
				Value:    value,
				DefValue: value.String(),
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

// Var defines a flag with the specified name and usage string. The type and
// value of the flag are represented by the first argument, of type Value, which
// typically holds a user-defined implementation of Value. For instance, the
// caller could create a flag that turns a comma-separated string into a slice
// of strings by giving the slice the methods of Value; in particular, Set would
// decompose the comma-separated string into the slice.
func Var(value Value, name string, usage string) {
	CommandLine.Var(value, name, usage)
}

// failf prints to standard error a formatted error and usage message and
// returns the error.
func (f *FlagSet) failf(format string, a ...interface{}) error {
	err := fmt.Errorf(format, a...)
	fmt.Fprintln(f.Output(), err)
	f.usage()
	return err
}

// usage calls the Usage method for the flag set if one is specified,
// or the appropriate default usage function otherwise.
func (f *FlagSet) usage() {
	if f.Usage == nil {
		f.defaultUsage()
	} else {
		f.Usage()
	}
}

func parseNameValue(name string) (string, bool, string) {
	for i := 1; i < len(name); i++ { // equals cannot be first
		if name[i] == '=' {
			return name[0:i], true, name[i+1:]
		}
	}

	return name, false, ""
}

func (f *FlagSet) getFlag(name string) (*Flag, bool, error) {
	formals := make([]map[string]*Flag, 0, 4)
	formals = append(formals, f.formal, f.shortLong, f.regex)

	if name == "help" || name == "h" { // special case for nice help message.
		f.usage()
		return nil, false, ErrHelp
	}

	if name == "version" || name == "V" {
		f.printVersion()
		return nil, false, ErrVersion
	}

	for k, formal := range formals {
		if k == len(formals)-1 && len(formal) > 0 { //range regexp map
			for k, v := range formal {
				//TODO: Compile and the full Regexp interface.
				matched, _ := regexp.Match(k, []byte(name))
				if matched {
					return v, true, nil
				}
			}
			continue
		}

		flag, alreadythere := formal[name] // BUG
		if !alreadythere {
			continue
		}

		return flag, true, nil
	}

	return nil, false, fmt.Errorf("flag provided but not defined: -%s", name)
}

func (f *FlagSet) setFlag(flag *Flag, name string, hasValue bool, value string) (bool, error) {

	if seen, err := f.setValue(flag, name, hasValue, value); err != nil {
		return seen, err
	}

	if f.actual == nil {
		f.actual = make(map[string]*Flag)
	}
	f.actual[name] = flag
	return true, nil
}

// 核心函数
func (f *FlagSet) setValue(flag *Flag, name string, hasValue bool, value string) (bool, error) {
	if flag.flags&NotValue > 0 {
		reflect.ValueOf(flag.pointer).Elem().Set(reflect.ValueOf(flag.matchValue))
		return true, nil
	}

	if fv, ok := flag.Value.(boolFlag); ok && fv.IsBoolFlag() { // special case: doesn't need an arg
		if hasValue {
			if err := fv.Set(value); err != nil {
				return false, f.failf("invalid boolean value %q for -%s: %v", value, name, err)
			}
		} else {
			if err := fv.Set("true"); err != nil {
				return false, f.failf("invalid boolean flag %s: %v", name, err)
			}
		}
		return true, nil
	}

	if fv, ok := flag.Value.(*boolSlice); ok {
		if err := fv.Set("true"); err != nil {
			return false, f.failf("invalid boolean flag %s: %v", name, err)
		}
		return true, nil
	}
	// It must have a value, which might be the next argument.
	if !hasValue && len(f.args) > 0 {
		// value is the next arg
		hasValue = true
		value, f.args = f.args[0], f.args[1:]
	}
	if !hasValue {
		return false, f.failf("flag needs an argument: -%s", name)
	}

	if err := flag.Value.Set(value); err != nil {
		return false, f.failf("invalid value %q for flag -%s: %v", value, name, err)
	}

	return true, nil
}

func (f *FlagSet) getName(numMinuses *int, name *string, flags Flags) (bool, bool, error) {
	if len(f.args) == 0 {
		return false, false, nil
	}
	s := f.args[0]
	if len(s) < 2 || s[0] != '-' {
		if flags&GreedyMode != GreedyMode {
			f.unkownArgs = append(f.unkownArgs, s)
		}
		f.args = f.args[1:]
		*name = s
		return false, true, nil
	}

	*numMinuses = 1
	if s[1] == '-' {
		(*numMinuses)++
		if len(s) == 2 { // "--" terminates the flags
			f.args = f.args[1:]
			return false, false, nil
		}
	}

	*name = s[*numMinuses:]
	if len(*name) == 0 || (*name)[0] == '-' || (*name)[0] == '=' {
		return false, false, f.failf("bad flag syntax: %s", s)
	}

	return true, true, nil
}

func (f *FlagSet) setPosixShortOpt(numMinuses int, name string) (bool, bool, error) {
	count := 0
	for i := range name {
		newName := string(name[i])

		flag, seen, err := f.getFlag(newName)
		if err != nil {
			continue
		}

		if (flag.flags & PosixShort) == 0 {
			continue
		}

		hasValue := false
		value := ""
		isBool := true

		// tail -c+3
		if fv, ok := flag.Value.(boolFlag); !(ok && fv.IsBoolFlag()) {

			if i+1 != len(name) {
				hasValue = true
				value = name[i+1:]
			}

			isBool = false
		}

		if flag.flags&RegexKeyIsValue > 0 {
			value = name[i:]
			hasValue = true
		}

		seen, err = f.setFlag(flag, newName, hasValue, value)
		if err != nil {
			return false, seen, err
		}

		if !isBool && !hasValue {
			return false, true, nil
		}

		count++
	}

	if count > 0 {
		return false, true, nil
	}

	return true, true, nil
}

func (f *FlagSet) setPosix(seen bool, err error, numMinuses int, name string) (bool, bool, error) {
	if err == ErrHelp {
		return false, false, err
	}

	if !f.openPosixShort {
		return false, seen, err
	}

	if numMinuses == 1 {
		next, seen, err := f.setPosixShortOpt(numMinuses, name)
		if !next {
			return next, seen, err
		}
	}
	return false, false, f.failf("%s", err.Error())
}

// parseOne parses one flag. It reports whether a flag was seen.
func (f *FlagSet) parseOne() (bool, error) {

	name, numMinuses := "", 0

	next, seen, err := f.getName(&numMinuses, &name, 0)
	if !next {
		return seen, err
	}
	// it's a flag. does it have an argument?
	f.args = f.args[1:]

	name, hasValue, value := parseNameValue(name)

	var (
		flag *Flag
		err0 error
	)

	if flag, seen, err0 = f.getFlag(name); err0 != nil {
		if next, seen, err0 = f.setPosix(seen, err0, numMinuses, name); !next {
			return seen, err0
		}
	}

try:
	if flag.flags&GreedyMode > 0 {
		for len(f.args) > 0 {

			name, numMinuses = "", 0

			seen, err = f.setFlag(flag, name, hasValue, value)
			if err != nil {
				return seen, err
			}
			next, seen, err = f.getName(&numMinuses, &name, GreedyMode)
			if !seen {
				return seen, err
			}
			if next {
				name, hasValue, value = parseNameValue(name)
				//fmt.Printf("---> name(%s), hasValue(%t), value(%s) args(%s)\n", name, hasValue, value, f.args)
				if flag, seen, err0 = f.getFlag(name); err0 != nil {
					if next, seen, err0 = f.setPosix(seen, err0, numMinuses, name); !next {
						return seen, err0
					}
				}

				f.args = f.args[1:]
				goto try
			}
			hasValue = true
			value = name
		}
	}

	if flag.flags&RegexKeyIsValue > 0 {
		value = name
		hasValue = true
	}
	return f.setFlag(flag, name, hasValue, value)

}

// Auto-generated Help, Version, and Usage information
func (f *FlagSet) generatedOpt() {
	f.Bool("h, help", false, "display this help and exit")
	f.Bool("V, version", false, "output version information and exit")
}

// Parse parses flag definitions from the argument list, which should not
// include the command name. Must be called after all flags in the FlagSet
// are defined and before flags are accessed by the program.
// The return value will be ErrHelp if -help or -h were set but not defined.
func (f *FlagSet) Parse(arguments []string) error {

	f.parsed = true
	f.args = arguments

	defer func() {
		f.args = append(f.unkownArgs, f.args...)
	}()

	for {
		seen, err := f.parseOne()
		if seen {
			continue
		}
		if err == nil {
			break
		}
		switch f.errorHandling {
		case ContinueOnError:
			return err
		case ExitOnError:
			os.Exit(2)
		case PanicOnError:
			panic(err)
		}
	}
	return nil
}

// Parsed reports whether f.Parse has been called.
func (f *FlagSet) Parsed() bool {
	return f.parsed
}

// Parse parses the command-line flags from os.Args[1:]. Must be called
// after all flags are defined and before flags are accessed by the program.
func Parse() {
	// Ignore errors; CommandLine is set for ExitOnError.
	CommandLine.Parse(os.Args[1:])
}

// Parsed reports whether the command-line flags have been parsed.
func Parsed() bool {
	return CommandLine.Parsed()
}

// CommandLine is the default set of command-line flags, parsed from os.Args.
// The top-level functions such as BoolVar, Arg, and so on are wrappers for the
// methods of CommandLine.
var CommandLine = NewFlagSet(os.Args[0], ExitOnError)

func init() {
	// Override generic FlagSet default Usage with call to global Usage.
	// Note: This is not CommandLine.Usage = Usage,
	// because we want any eventual call to use any updated value of Usage,
	// not the value it has when this line is run.
	CommandLine.Usage = commandLineUsage
}

func commandLineUsage() {
	Usage()
}

// NewFlagSet returns a new, empty flag set with the specified name and
// error handling property.
func NewFlagSet(name string, errorHandling ErrorHandling) *FlagSet {
	f := &FlagSet{
		name:          name,
		errorHandling: errorHandling,
	}
	f.Usage = f.defaultUsage
	f.generatedOpt()
	return f
}

//
//
func (f *FlagSet) Version(version string) *FlagSet {
	f.version = version
	return f
}

func (f *FlagSet) Author(author string) *FlagSet {
	f.author = author
	return f
}

func Version(version string) {
	CommandLine.Version(version)
}

func Author(author string) {
	CommandLine.Author(author)
}

// Init sets the name and error handling property for a flag set.
// By default, the zero FlagSet uses an empty name and the
// ContinueOnError error handling policy.
func (f *FlagSet) Init(name string, errorHandling ErrorHandling) {
	f.name = name
	f.errorHandling = errorHandling
}
