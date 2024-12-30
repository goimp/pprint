package pprint

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

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

func recursion(object interface{}) string {
	// Get the type of the object
	objectType := reflect.TypeOf(object).Name()

	// Get the memory address (simulating id in Python)
	// Use reflect.ValueOf to get the value, then convert to unsafe.Pointer
	// objectValue := reflect.ValueOf(object)
	// objectID := fmt.Sprintf("%v", unsafe.Pointer(objectValue.Pointer()))
	objectId := id(object)

	// Return the formatted string similar to Python's recursion function
	return fmt.Sprintf("<Recursion on %s with id=%v>", objectType, objectId)
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
		if len(fmt.Sprintf("%v", candidate)) > width {
			if len(current) > 0 {
				result = append(result, fmt.Sprintf("%v", current))
			}
			current = part
		} else {
			current = candidate
		}
	}

	// Add the final part if there's any remaining
	if len(current) > 0 {
		result = append(result, fmt.Sprintf("%v", current))
	}

	return result
}
