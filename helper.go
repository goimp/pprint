package pprint

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func repr(object any) string {
	// if _, p := ptrToObj(object); p {
	// 	// fmt.Println("PTR", object, o)
	// 	return fmt.Sprintf("%v", reflect.TypeOf(object))
	// }
	return fmt.Sprintf("%#v", object)
}

func ptrToObj(object any) (reflect.Value, bool) {
	value := reflect.ValueOf(object)
	if value.Kind() == reflect.Pointer {
		return reflect.Indirect(value), true
	}
	return value, false
}

func id(object any) uintptr {
	// Get the reflect.Value of the object
	value := reflect.ValueOf(object)

	// Check if the object is addressable (has a memory address)
	if value.Kind() == reflect.Ptr || value.CanAddr() {
		// Use uintptr to represent the memory address
		return value.Pointer()
	}

	// If not addressable, create a pointer to the object to get its address
	return reflect.ValueOf(&object).Pointer()
}

func getType[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

func copyContext(context Context) Context {
	// Create a new map to store the copied data
	copy := make(Context, len(context))

	// Copy each key-value pair
	for key, value := range context {
		copy[key] = value
	}

	return copy
}

func formatWithUnderscores(num int) string {
	// Convert the integer to a string
	numStr := strconv.Itoa(num)

	// Add underscores for thousands separators
	var sb strings.Builder
	length := len(numStr)
	for i, digit := range numStr {
		sb.WriteRune(digit)
		// Add an underscore after every 3 digits (except at the end)
		if (length-i-1)%3 == 0 && i != length-1 {
			sb.WriteRune('_')
		}
	}
	return sb.String()
}

func recursion(object any) string {
	objectType := reflect.TypeOf(object).Name()
	objectId := id(object)
	return fmt.Sprintf("<Recursion on %s with id=%s>", objectType, repr(objectId))
}

func wrapBytesRepr(object []byte, width, allowance int) []string {
	var result []string
	var current []byte
	last := len(object) / 4 * 4

	for i := 0; i < len(object); i += 4 {
		part := object[i : i+4]
		candidate := append(current, part...)
		if i == last {
			width -= allowance
		}
		if len(repr(candidate)) > width {
			if len(current) > 0 {
				result = append(result, repr(current))
			}
			current = part
		} else {
			current = candidate
		}
	}

	// Add the final part if there's any remaining
	if len(current) > 0 {
		result = append(result, repr(current))
	}

	return result
}
