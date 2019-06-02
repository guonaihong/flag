package flag

import (
	"os"
	"reflect"
	"strconv"
	"strings"
)

func parseFlags(s string) (f Flags) {

	fs := strings.Split(s, "|")
	for _, v := range fs {
		switch v {
		case "posix":
			f |= PosixShort
		case "greedy":
			f |= GreedyMode
		}
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

		if v.Type() == stringSliceType {
			rv = strings.Split(defValue, sep)
		} else if v.Type() == int64SliceType {
			rs := strings.Split(defValue, sep)
			int64s := make([]int64, len(rs))
			for k, v := range rs {
				i64, err := strconv.ParseInt(v, 10, 0)
				if err != nil {
					panic(err.Error())
				}
				int64s[k] = i64
			}
		} else {
			panic("unkown slice type")
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
		rv, err = strconv.ParseInt(defValue, 10, 0)
	case reflect.Float64:
		rv, err = strconv.ParseFloat(defValue, 0)
	case reflect.Bool:
		rv, err = strconv.ParseBool(defValue)
	case reflect.String:
		rv = defValue
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
