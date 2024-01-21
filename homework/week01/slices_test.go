package week01

import (
	"reflect"
	"testing"
)

func TestDeleteByIndex(t *testing.T) {
	src1 := []int{1, 2, 3, 4, 5}
	res1, error := DeleteByIndex[int](src1, 3)
	if error != nil {
		t.Error(error)
	}
	equal := reflect.DeepEqual(res1, []int{1, 2, 3, 5})
	if !equal {
		t.Errorf("not equal")
	}

	src2 := []bool{true, false, false}
	res2, error := DeleteByIndex[bool](src2, 1)
	if error != nil {
		t.Error(error)
	}
	equal2 := reflect.DeepEqual(res2, []bool{true, false})
	if !equal2 {
		t.Errorf("not equal")
	}

}

func TestDeleteByIndexOfCopy(t *testing.T) {
	src1 := []int{1, 2, 3, 4, 5}
	res1, error := DeleteByIndexOfCopy[int](src1, 3)
	if error != nil {
		t.Error(error)
	}
	equal1 := reflect.DeepEqual(res1, []int{1, 2, 3, 5})
	if !equal1 {
		t.Errorf("not equal")
	}

	src2 := []bool{true, false, false}
	res2, error := DeleteByIndexOfCopy[bool](src2, 1)
	if error != nil {
		t.Error(error)
	}
	equal2 := reflect.DeepEqual(res2, []bool{true, false})
	if !equal2 {
		t.Errorf("not equal")
	}

}
