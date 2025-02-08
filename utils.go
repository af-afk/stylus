package stylus

import "reflect"

//go:inline
func unsafePtr(x any) int32 {
	return *(*int32)(reflect.ValueOf(x).UnsafePointer())
}