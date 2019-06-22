package flag

import (
	"bytes"
	"fmt"
	"reflect"
	"time"
)

var boolSliceType = reflect.TypeOf([]bool{})

var stringSliceType = reflect.TypeOf([]string{})

var int64SliceType = reflect.TypeOf([]int64{})

var durationType = reflect.TypeOf(time.Duration(1))

func (f *FlagSet) setNamesToMap(m *map[string]*Flag, names []string, flag *Flag) {

	initFormal(m)
	for _, v := range names {
		_, alreadythere := (*m)[v]
		if alreadythere {
			f.alreadythereError(v)
		}

		newFlag := *flag
		newFlag.Name = v
		(*m)[v] = &newFlag
	}
}

func (f *FlagSet) flagVar(flag *Flag) {

	if flag.flags&PosixShort > 0 && flag.flags&GreedyMode > 0 {
		panic("Cannot set both PosixShort and GreedyMode")
	}

	name := flag.Name
	var names []string
	var ok bool

	if flag.isOptOpt {
		f.setNamesToMap(&f.regex, []string{flag.Regex}, flag)
		flag.flags ^= RegexKeyIsValue
		f.setNamesToMap(&f.shortLong, flag.Short, flag)
		f.setNamesToMap(&f.shortLong, flag.Long, flag)
		name = flag.Name
	} else {
		name, names, ok = newName(name)
		if ok {
			f.setNamesToMap(&f.shortLong, names, flag)
		}
		flag.Name = name
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
		opt.isOptOpt = true
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

type InvalidVarError struct {
	Type reflect.Type
}

func (e *InvalidVarError) Error() string {
	if e.Type == nil {
		return "flag: Var(nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "flag: Var(non-pointer " + e.Type.String() + ")"
	}

	return "flag: Var(nil " + e.Type.String() + ")"
}

func (f *Flag) setVar(defValue, p reflect.Value) {
	vt := p.Elem().Type()
	v := p.Elem().Type()

	switch v.Kind() {
	case reflect.Uint8:
		if vt == reflect.TypeOf(byte(0)) {
			f.Value = newByteValue(defValue.Interface().(byte), p.Interface().(*byte))
		} else {
			panic("unkown type")
		}
	case reflect.String:
		f.Value = newStringValue(defValue.Interface().(string), p.Interface().(*string))
	case reflect.Bool:
		f.Value = newBoolValue(defValue.Interface().(bool), p.Interface().(*bool))
	case reflect.Uint:
		f.Value = newUintValue(defValue.Interface().(uint), p.Interface().(*uint))
	case reflect.Uint64:
		f.Value = newUint64Value(defValue.Interface().(uint64), p.Interface().(*uint64))
	case reflect.Int:
		f.Value = newIntValue(defValue.Interface().(int), p.Interface().(*int))
	case reflect.Int64:
		if durationType == vt {
			f.Value = newDurationValue(defValue.Interface().(time.Duration), p.Interface().(*time.Duration))
		} else {
			f.Value = newInt64Value(defValue.Interface().(int64), p.Interface().(*int64))
		}
	case reflect.Float64:
		f.Value = newFloat64Value(defValue.Interface().(float64), p.Interface().(*float64))
	case reflect.Slice:
		switch vt {
		case stringSliceType:
			f.Value = newStringSliceValue(defValue.Interface().([]string), p.Interface().(*[]string))
		case int64SliceType:
			f.Value = newInt64SliceValue(defValue.Interface().([]int64), p.Interface().(*[]int64))
		case boolSliceType:
			f.Value = newBoolSliceValue(defValue.Interface().([]bool), p.Interface().(*[]bool))
		default:
			panic(fmt.Sprintf("%v:Unsupported type", vt))
		}
	default:
		panic("unkown type")
	}

	f.parent.flagVar(f)
}

func checkValue(p, defValue interface{}, rv reflect.Value) {
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		p := &InvalidVarError{reflect.TypeOf(p)}
		panic(p.Error())
	}

	if reflect.TypeOf(defValue) != rv.Elem().Type() {
		panic(fmt.Sprintf("defvalue type is %v: value type is %v\n",
			reflect.TypeOf(defValue), rv.Elem().Type()))
	}
}

func (f *Flag) Var(p interface{}) {
	rv := reflect.ValueOf(p)

	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		p := &InvalidVarError{reflect.TypeOf(p)}
		panic(p.Error())
	}

	f.setVar(reflect.Zero(rv.Elem().Type()), rv)
}

func (f *Flag) DefaultVar(p, defValue interface{}) {
	rv := reflect.ValueOf(p)

	checkValue(p, defValue, rv)

	f.setVar(reflect.ValueOf(defValue), rv)
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

func (f *Flag) NewBoolSlice(defValue []bool) *[]bool {
	p := new([]bool)
	f.Value = newBoolSliceValue(defValue, p)
	f.parent.flagVar(f)
	return p
}

func Opt(name string, usage string) *Flag {
	return CommandLine.Opt(name, usage)
}
