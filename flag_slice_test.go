package flag

import (
	"testing"
)

func intSliceCmp(int64Slice0, int64Slice1 []int64) bool {
	if len(int64Slice0) != len(int64Slice1) {
		return false
	}

	for k, _ := range int64Slice0 {
		if int64Slice0[k] != int64Slice1[k] {
			return false
		}
	}

	return true
}

func stringSliceCmp(stringSlice0, stringSlice1 []string) bool {
	if len(stringSlice0) != len(stringSlice1) {
		return false
	}

	for k, _ := range stringSlice0 {
		if stringSlice0[k] != stringSlice1[k] {
			return false
		}
	}

	return true
}
func testSliceParse(f *FlagSet, t *testing.T) {
	intSliceFlag := f.Int64Slice("int64slice", []int64{}, "int 64 slice")
	stringSliceFlag := f.StringSlice("stringslice", []string{}, "string slice")
	args := []string{
		"--int64slice", "1",
		"--int64slice", "2",
		"--int64slice", "3",
		"--int64slice", "4",
		"--stringslice", "header1",
		"--stringslice", "header2",
		"--stringslice", "header3",
	}

	if err := f.Parse(args); err != nil {
		t.Fatal(err)
	}

	needIntSliceValue := []int64{1, 2, 3, 4}
	if !intSliceCmp(*intSliceFlag, needIntSliceValue) {
		t.Errorf("int slice falg shout be %v, is %v ", needIntSliceValue, intSliceFlag)
	}

	needStringSliceValue := []string{"header1", "header2", "header3"}
	if !stringSliceCmp(*stringSliceFlag, needStringSliceValue) {
		t.Errorf("string slice falg shout be %v, is %v ", needStringSliceValue, stringSliceFlag)
	}
}

func TestSliceParse(t *testing.T) {
	testSliceParse(CommandLine, t)
}
