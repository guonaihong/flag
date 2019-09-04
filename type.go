package flag

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

// -- bool Value
type boolValue bool

func newBoolValue(val bool, p *bool) *boolValue {
	*p = val
	return (*boolValue)(p)
}

func (b *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	*b = boolValue(v)
	return err
}

func (b *boolValue) Get() interface{} { return bool(*b) }

func (b *boolValue) String() string { return strconv.FormatBool(bool(*b)) }

func (b *boolValue) IsBoolFlag() bool { return true }

// optional interface to indicate boolean flags that can be
// supplied without "=value" text
type boolFlag interface {
	Value
	IsBoolFlag() bool
}

// -- byte value
type byteValue byte

func newByteValue(val byte, p *byte) *byteValue {
	*p = val
	return (*byteValue)(p)
}

func (b *byteValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 10, 8)
	*b = byteValue(byte(v))
	return err
}

func (b *byteValue) Get() interface{} { return byte(*b) }

func (b *byteValue) String() string { return strconv.Itoa(int(*b)) }

// -- int Value
type intValue int

func newIntValue(val int, p *int) *intValue {
	*p = val
	return (*intValue)(p)
}

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 10, 32)
	*i = intValue(v)
	return err
}

func (i *intValue) Get() interface{} { return int(*i) }

func (i *intValue) String() string { return strconv.Itoa(int(*i)) }

// -- int64 Value
type int64Value int64

func newInt64Value(val int64, p *int64) *int64Value {
	*p = val
	return (*int64Value)(p)
}

func (i *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*i = int64Value(v)
	return err
}

func (i *int64Value) Get() interface{} { return int64(*i) }

func (i *int64Value) String() string { return strconv.FormatInt(int64(*i), 10) }

// -- uint Value
type uintValue uint

func newUintValue(val uint, p *uint) *uintValue {
	*p = val
	return (*uintValue)(p)
}

func (i *uintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, strconv.IntSize)
	*i = uintValue(v)
	return err
}

func (i *uintValue) Get() interface{} { return uint(*i) }

func (i *uintValue) String() string { return strconv.FormatUint(uint64(*i), 10) }

// -- uint64 Value
type uint64Value uint64

func newUint64Value(val uint64, p *uint64) *uint64Value {
	*p = val
	return (*uint64Value)(p)
}

func (i *uint64Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	*i = uint64Value(v)
	return err
}

func (i *uint64Value) Get() interface{} { return uint64(*i) }

func (i *uint64Value) String() string { return strconv.FormatUint(uint64(*i), 10) }

// -- string Value
type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Get() interface{} { return string(*s) }

func (s *stringValue) String() string { return string(*s) }

// -- float64 Value
type float64Value float64

func newFloat64Value(val float64, p *float64) *float64Value {
	*p = val
	return (*float64Value)(p)
}

func (f *float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	*f = float64Value(v)
	return err
}

func (f *float64Value) Get() interface{} { return float64(*f) }

func (f *float64Value) String() string { return strconv.FormatFloat(float64(*f), 'g', -1, 64) }

// -- time.Duration Value
type durationValue time.Duration

func newDurationValue(val time.Duration, p *time.Duration) *durationValue {
	*p = val
	return (*durationValue)(p)
}

func (d *durationValue) Set(s string) error {
	v, err := time.ParseDuration(s)
	*d = durationValue(v)
	return err
}

func (d *durationValue) Get() interface{} { return time.Duration(*d) }

func (d *durationValue) String() string { return (*time.Duration)(d).String() }

// -- duration slice value
type durationSliceValue []time.Duration

func newDurationSliceValue(val []time.Duration, p *[]time.Duration) *durationSliceValue {
	*p = val
	return (*durationSliceValue)(p)
}

func (d *durationSliceValue) Set(val string) error {
	var dv durationValue

	err := dv.Set(val)
	if err != nil {
		return err
	}

	*d = append(*d, time.Duration(dv))
	return nil
}

func (d *durationSliceValue) Get() interface{} {
	return []time.Duration(*d)
}

func (d *durationSliceValue) String() string {
	switch len(*d) {
	case 0:
		return "[]"
	case 1:
		return fmt.Sprintf("[%s]", (*d)[0])
	case 2:
		return fmt.Sprintf("[%s, %s]", (*d)[0], (*d)[1])
	case 3:
		return fmt.Sprintf("[%s, %s, %s]", (*d)[0], (*d)[1], (*d)[2])
	}

	var buf bytes.Buffer

	buf.WriteString("[")

	for k, v := range *d {
		buf.WriteString((*durationValue)(&v).String())
		if k != len(*d)-1 {
			buf.WriteString(",")
		}
	}

	buf.WriteString("]")

	return buf.String()
}
