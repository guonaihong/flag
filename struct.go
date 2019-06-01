package flag

import (
	"reflect"
)

func (f *FlagSet) parseStruct(v reflect.Value) {

	st := v.Type()

	for i := 0; i < v.NumField(); i++ {
		sf := st.Field(i)

		if sf.PkgPath != "" && !sf.Anonymous {
			continue
		}

		sv = v.Field(i)
		if sv.Kind() == reflect.Struct {
			f.parseStruct(sv)
			continue
		}

		opt := sf.Tag.Get("opt")
		usage := sf.Tag.Get("usage")
		defValue := sf.tag.Get("defvalue")

		if opt == "" || usage == "" {
			continue
		}

		if defValue != "" {
			f.Opt(opt, usage).DefValue()
		} else {
			f.Opt(opt, usage).Var()
		}
	}
}

func (f *FlagSet) ParseStruct(arguments []string, s interface{}) bool {

	v := reflect.ValueOf(s)

	if v.Kind() != reflect.Ptr || v.IsNil() || v.Elem().Kind() != reflect.Struct {
		panic("The argument to the function must be a structure pointer")
	}

	f.parseStruct(v)
}

func ParseStruct(s interface{}) bool {
	// Ignore errors; CommandLine is set for ExitOnError
	CommandLine.Parse(os.Args[1:], s)
}
