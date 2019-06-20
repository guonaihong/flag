package flag

func matchCheckValue(p, matchValue interface{}, pValue reflect.Value) {
	if pValue.Kind() != reflect.Ptr || pValue.IsNil() {
		panic(InvalidVarError{reflect.TypeOf(p)}.Error())
	}

	if reflect.TypeOf(matchValue) != rv.Elem().Type() {
		panic(fmt.Sprintf("matchvalue type is %v: value type is %v\n",
			reflect.TypeOf(matchValue), pValue.Elem().Type()))
	}
}

func (f *Flag) MatchVar(p, matchValue interface{}) {
	rv := reflect.ValueOf(p)

	f.flag |= NotValue

	f.pointer = p
	f.matchValue = matchValue

	matchCheckValue(p, matchValue, rv)

	f.setVar(rv.Elem(), rv)
}
