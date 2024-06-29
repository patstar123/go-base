package utils

import (
	"reflect"
	"unsafe"
)

func IsSliceAddressEqual(slice1, slice2 []byte) bool {
	head1 := (*reflect.SliceHeader)(unsafe.Pointer(&slice1))
	head2 := (*reflect.SliceHeader)(unsafe.Pointer(&slice2))
	return head1.Data == head2.Data && head1.Cap == head2.Cap && head1.Len == head2.Len
}

func IsSliceAddressEqual2(slice1, slice2 []byte) bool {
	return reflect.DeepEqual((*reflect.SliceHeader)(unsafe.Pointer(&slice1)), (*reflect.SliceHeader)(unsafe.Pointer(&slice2)))
}
