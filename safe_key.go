package pprint

import (
	"fmt"
	"reflect"
	"unsafe"
)

type safeKey struct {
	obj any
}

func newSafeKey(obj any) safeKey {
	return safeKey{obj: obj}
}

// lessThan compares two safeKey objects.
func (sk safeKey) lessThan(other safeKey) bool {
	// Attempt direct comparison if possible
	if reflect.TypeOf(sk.obj) == reflect.TypeOf(other.obj) {
		switch v := sk.obj.(type) {
		case int:
			return v < other.obj.(int)
		case string:
			return v < other.obj.(string)
		// Add cases for other directly comparable types as needed
		default:
			// If types are the same but not directly comparable, fall back
			return sk.fallbackCompare(other)
		}
	}

	// Fallback for unorderable or mismatched types
	return sk.fallbackCompare(other)
}

// fallbackCompare handles fallback comparison logic.
func (sk safeKey) fallbackCompare(other safeKey) bool {
	// Compare types first
	typeStr1 := fmt.Sprintf("%T", sk.obj)
	typeStr2 := fmt.Sprintf("%T", other.obj)
	if typeStr1 != typeStr2 {
		return typeStr1 < typeStr2
	}

	// If types are the same, compare memory addresses
	return uintptr(unsafe.Pointer(&sk.obj)) < uintptr(unsafe.Pointer(&other.obj))
}

// Helper function for comparing 2-slices
func safeTuple(t []any) (safeKey, safeKey) {
	return safeKey{obj: t[0]}, safeKey{obj: t[1]}
}
