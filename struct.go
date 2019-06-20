package flag

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func parseFlags(s string) (f Flags) {

	fs := strings.Split(s, "|")
	for _, v := range fs {
		switch v {
		case "posix", "Posix":
			f |= PosixShort
		case "greedy", "Greedy":
			f |= GreedyMode
		case "notValue", "NotValue":
			f |= NotValue
		}
	}
	return
}

func parseByte(s string) (b byte, err error) {

	switch s {
	case `\a`:
		b = '\a'
	case `\b`:
		b = '\b'
	case `\e`:
		b = '\x1B'
	case `\f`:
		b = '\f'
	case `\n`:
		b = '\n'
	case `\r`:
		b = '\r'
	case `\v`:
		b = '\v'

	default:
		var u64 uint64
		switch {
		case strings.HasPrefix(s, "x"):
			if u64, err = strconv.ParseUint(s[1:], 16, 16); err != nil {
				return 0, err
			}

		case strings.HasPrefix(s, "0"):
			if u64, err = strconv.ParseUint(s[1:], 8, 16); err != nil {
				return 0, err
			}
		default:
			if u64, err = strconv.ParseUint(s[1:], 10, 16); err != nil {
				return 0, err
			}

		}

		b = byte(u64)
	}

	return

}

func parseDefValue(v reflect.Value, defValue string, sep string) (rv interface{}) {
	var err error
	switch v.Kind() {
	case reflect.Slice:
		if sep == "" {
			sep = ","
		}

		switch v.Type() {
		case stringSliceType:
			rv = strings.Split(defValue, sep)
		case int64SliceType:
			rs := strings.Split(defValue, sep)
			int64s := make([]int64, len(rs))
			for k, v := range rs {
				i64, err := strconv.ParseInt(v, 10, 0)
				if err != nil {
					panic(err.Error())
				}
				int64s[k] = i64
			}
			rv = int64s
		default:
			panic(fmt.Sprintf("unkown slice type:%v #support []stirng and []int64 types", v.Type()))
		}

	case reflect.Uint, reflect.Uint64:
		n := uint64(0)
		n, err = strconv.ParseUint(defValue, 10, 0)
		rv = n
		if v.Kind() == reflect.Uint {
			rv = uint(n)
		}

	case reflect.Int:
		rv, err = strconv.Atoi(defValue)
	case reflect.Int64:
		if v.Type() == durationType {
			rv, err = time.ParseDuration(defValue)
		} else {
			rv, err = strconv.ParseInt(defValue, 10, 0)
		}
	case reflect.Float64:
		rv, err = strconv.ParseFloat(defValue, 0)
	case reflect.Bool:
		rv, err = strconv.ParseBool(defValue)
	case reflect.String:
		rv = defValue
	case reflect.Uint8:
		if reflect.TypeOf(byte(0)) == v.Type() {
			rv, err = parseByte(defValue)
		} else {
			panic("invalid type")
		}
	default:
		panic("invalid type")
	}

	if err != nil {
		panic(err.Error())
	}

	return
}

func (f *FlagSet) parseStruct(v reflect.Value) bool {

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	st := v.Type()

	for i := 0; i < st.NumField(); i++ {
		sf := st.Field(i)

		if sf.PkgPath != "" && !sf.Anonymous {
			continue
		}

		sv := v.Field(i)
		if sv.Kind() == reflect.Struct {
			f.parseStruct(sv)
			continue
		}

		opt := sf.Tag.Get("opt")
		usage := sf.Tag.Get("usage")
		defValue := sf.Tag.Get("defValue")
		flags := sf.Tag.Get("flags")

		if opt == "" || usage == "" {
			continue
		}

		if defValue != "" {
			f.Opt(opt, usage).
				Flags(parseFlags(flags)).
				DefaultVar(sv.Addr().Interface(), parseDefValue(sv, defValue, sf.Tag.Get("sep")))
		} else {
			f.Opt(opt, usage).
				Flags(parseFlags(flags)).
				Var(sv.Addr().Interface())
		}
	}
	return true
}

func (f *FlagSet) ParseStruct(arguments []string, s interface{}) error {

	v := reflect.ValueOf(s)

	if v.Kind() != reflect.Ptr || v.IsNil() || v.Elem().Kind() != reflect.Struct {
		panic("The argument to the function must be a structure pointer")
	}

	f.parseStruct(v)

	return f.Parse(arguments)
}

func ParseStruct(s interface{}) {
	// Ignore errors; CommandLine is set for ExitOnError
	CommandLine.ParseStruct(os.Args[1:], s)
}
