// -------------------------------------------------------------------------------------------------------

package pprint

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type Serializer func(val reflect.Value, includePrivateFields bool) any

type Marshalizer struct {
	context map[uintptr]string
}

func SerializePointer(val reflect.Value, includePrivateFields bool) any {
	// Handle pointer types
	if val.IsNil() {
		return nil
	}

	// Create a structure for pointers
	ptrData := map[string]any{
		"address": fmt.Sprintf("%p", val.Interface()), // Pointer address
		// "address": id(val.Interface()),                // Pointer address
		"type": fmt.Sprintf("%T", val.Interface()), // Pointer type
	}

	// Serialize the dereferenced value first
	value := serializeUnsupported(val.Elem().Interface(), includePrivateFields)

	// If the dereferenced value is a map (structure-like), add pointer data
	if reflect.ValueOf(value).Kind() == reflect.Map {
		// Add pointer metadata to the map
		value.(map[string]any)["*"] = ptrData
		return value
	}
	// If it's not a map, just return the value as is
	return value
}

func SerializeStruct(val reflect.Value, includePrivateFields bool) any {
	// Handle structs
	m := make(map[string]any)
	typ := val.Type()
	// typ := reflect.TypeOf(object)
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		if !field.CanInterface() {
			// optional
			if includePrivateFields {
				m[fieldType.Name] = "[private]"
			}
			continue // Skip unexported fields
		}
		m[fieldType.Name] = serializeUnsupported(field.Interface(), includePrivateFields)
	}
	return m
}

func SerializeSlice(val reflect.Value, includePrivateFields bool) any {
	// Handle slices
	result := make([]any, val.Len())
	for i := 0; i < val.Len(); i++ {
		result[i] = serializeUnsupported(val.Index(i).Interface(), includePrivateFields)
	}
	return result
}

func SerializeMap(val reflect.Value, includePrivateFields bool) any {
	// Handle maps
	result := make(map[string]any)
	for _, key := range val.MapKeys() {
		result[fmt.Sprintf("%v", key.Interface())] = serializeUnsupported(val.MapIndex(key).Interface(), includePrivateFields)
	}
	return result
}

func SerializeFuncSignature(val reflect.Value, includePrivateFields bool) any {
	// Automatically generate a function descriptor
	funcType := val.Type()

	// Get parameter types of the function (if any)
	numParams := funcType.NumIn()
	paramTypes := []string{}
	for i := 0; i < numParams; i++ {
		if typestring := funcType.In(i).String(); len(typestring) > 0 {
			paramTypes = append(paramTypes, removeSpaces(typestring))
		}
	}

	funcName := funcType.Name()
	if len(funcName) > 0 {
		funcName += " "
	}

	numReturns := funcType.NumOut()
	returnTypes := []string{}
	for i := 0; i < numReturns; i++ {
		if typestring := funcType.Out(i).String(); len(typestring) > 0 {
			paramTypes = append(paramTypes, removeSpaces(typestring))
		}
	}
	
	fmt.Println(joinTypes(paramTypes), joinTypes(returnTypes))
	if len(returnTypes) == 0 {
		return fmt.Sprintf("%sfunc(%s)", funcName, joinTypes(paramTypes))
	} else if len(returnTypes) == 1 {
		return fmt.Sprintf("%sfunc(%s) %s", funcName, joinTypes(paramTypes), joinTypes(returnTypes))
	} else {
		return fmt.Sprintf("%sfunc(%s) (%s)", funcName, joinTypes(paramTypes), joinTypes(returnTypes))
	}
}

// CustomMarshal replaces unsupported types with string descriptors and controls HTML escaping.
func CustomMarshal(object any, escapeHTML bool, includePrivateFields bool) ([]byte, error) {
	// Marshal data with custom serialization
	serializedData := serializeUnsupported(object, includePrivateFields)

	// Marshal data with indentation and optional HTML escaping
	var result []byte
	var err error
	if escapeHTML {
		// If escapeHTML is true, use standard MarshalIndent
		result, err = json.MarshalIndent(serializedData, "", "  ")
	} else {
		// If escapeHTML is false, use encoder with SetEscapeHTML(false)
		encoder := json.NewEncoder(nil)
		encoder.SetEscapeHTML(false)
		result, err = json.MarshalIndent(serializedData, "", "  ")
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

// serializeUnsupported replaces unsupported types like functions with string descriptors.
func serializeUnsupported(object any, includePrivateFields bool) any {
	if object == nil {
		return nil
	}

	val := reflect.ValueOf(object)
	switch val.Kind() {
	case reflect.Pointer:
		return SerializePointer(val, includePrivateFields)
	case reflect.Struct:
		return SerializeStruct(val, includePrivateFields)
	case reflect.Slice:
		return SerializeSlice(val, includePrivateFields)
	case reflect.Map:
		return SerializeMap(val, includePrivateFields)
	case reflect.Func:
		return SerializeFuncSignature(val, includePrivateFields)
	default:
		// Return the value directly for supported types
		return object
	}
}

// joinTypes is a helper function to join type strings with commas.
func joinTypes(types []string) string {
	if len(types) == 0 {
		return ""
	}
	return strings.Join(types, ", ")
}

// Function to clean up spaces between type names
func removeSpaces(typeStr string) string {
	return strings.ReplaceAll(typeStr, " ", "")
}

// // ---------------------------------------------------------------------------------------------------------

// package pprint

// import (
// 	"encoding/json"
// 	"fmt"
// 	"reflect"
// )

// // CustomMarshal replaces unsupported types with string descriptors and controls HTML escaping.
// func CustomMarshal(object any, escapeHTML bool, includePrivateFields bool) ([]byte, error) {
// 	// Marshal data with custom serialization
// 	serializedData := serializeUnsupported(v, includePrivateFields)

// 	// Marshal data with indentation and optional HTML escaping
// 	var result []byte
// 	var err error
// 	if escapeHTML {
// 		// If escapeHTML is true, use standard MarshalIndent
// 		result, err = json.MarshalIndent(serializedData, "", "  ")
// 	} else {
// 		// If escapeHTML is false, use encoder with SetEscapeHTML(false)
// 		encoder := json.NewEncoder(nil)
// 		encoder.SetEscapeHTML(false)
// 		result, err = json.MarshalIndent(serializedData, "", "  ")
// 	}

// 	if err != nil {
// 		return nil, err
// 	}

// 	return result, nil
// }

// // serializeUnsupported replaces unsupported types like functions with string descriptors.
// func serializeUnsupported(object any, includePrivateFields bool) any {
// 	if v == nil {
// 		return nil
// 	}

// 	val := reflect.ValueOf(v)
// 	switch val.Kind() {
// 	case reflect.Pointer:
// 		// Dereference the pointer
// 		if val.IsNil() {
// 			return nil
// 		}
// 		// Handle pointers by marking them with a * prefix on the struct key
// 		result := serializeUnsupported(val.Elem().Interface(), includePrivateFields)
// 		// If it's a map, add "*" prefix to the struct key only
// 		if reflect.TypeOf(result).Kind() == reflect.Map {
// 			m := result.(map[string]any)
// 			for key := range m {
// 				// Only add * to the top-level key for the struct
// 				if key != "*private" { // To avoid adding * to private fields
// 					m["*"+key] = m[key] // Add '*' prefix to the top-level key
// 					delete(m, key)       // Remove the old key
// 				}
// 			}
// 		}
// 		return result
// 	case reflect.Struct:
// 		// Handle structs
// 		m := make(map[string]any)
// 		typ := reflect.TypeOf(v)
// 		for i := 0; i < val.NumField(); i++ {
// 			field := val.Field(i)
// 			fieldType := typ.Field(i)
// 			if !field.CanInterface() {
// 				// optional
// 				if includePrivateFields {
// 					m[fieldType.Name] = "[private]"
// 				}
// 				continue // Skip unexported fields
// 			}
// 			m[fieldType.Name] = serializeUnsupported(field.Interface(), includePrivateFields)
// 		}
// 		return m
// 	case reflect.Slice:
// 		// Handle slices
// 		result := make([]any, val.Len())
// 		for i := 0; i < val.Len(); i++ {
// 			result[i] = serializeUnsupported(val.Index(i).Interface(), includePrivateFields)
// 		}
// 		return result
// 	case reflect.Map:
// 		// Handle maps
// 		result := make(map[string]any)
// 		for _, key := range val.MapKeys() {
// 			result[fmt.Sprintf("%v", key.Interface())] = serializeUnsupported(val.MapIndex(key).Interface(), includePrivateFields)
// 		}
// 		return result
// 	default:
// 		if val.Kind() == reflect.Func {
// 			// Automatically generate a function descriptor
// 			funcType := val.Type()
// 			// funcName := funcType.Name()

// 			// Get parameter types of the function (if any)
// 			numParams := funcType.NumIn()
// 			paramTypes := make([]string, numParams)
// 			for i := 0; i < numParams; i++ {
// 				paramTypes[i] = funcType.In(i).String()
// 			}

// 			// Get return type if available
// 			var returnType string
// 			if funcType.NumOut() > 0 {
// 				returnType = funcType.Out(0).String()
// 			}

// 			// Create a descriptor string that includes the function name, parameter types, and return type if any
// 			if returnType != "" {
// 				return fmt.Sprintf("func(%s) %s", joinTypes(paramTypes), returnType)
// 			} else {
// 				return fmt.Sprintf("func(%s)", joinTypes(paramTypes))
// 			}
// 		}
// 		// Return the value directly for supported types
// 		return v
// 	}
// }

// // joinTypes is a helper function to join type strings with commas.
// func joinTypes(types []string) string {
// 	if len(types) == 0 {
// 		return ""
// 	}
// 	return fmt.Sprintf("%s", types)
// }

// package pprint

// import (
// 	"encoding/json"
// 	"fmt"
// 	"reflect"
// )

// // CustomMarshal replaces unsupported types with string descriptors and controls HTML escaping.
// func CustomMarshal(object any, escapeHTML bool, includePrivateFields bool) ([]byte, error) {
// 	// Marshal data with custom serialization
// 	serializedData := serializeUnsupported(v, includePrivateFields)

// 	// Marshal data with indentation and optional HTML escaping
// 	var result []byte
// 	var err error
// 	if escapeHTML {
// 		// If escapeHTML is true, use standard MarshalIndent
// 		result, err = json.MarshalIndent(serializedData, "", "  ")
// 	} else {
// 		// If escapeHTML is false, use encoder with SetEscapeHTML(false)
// 		encoder := json.NewEncoder(nil)
// 		encoder.SetEscapeHTML(false)
// 		result, err = json.MarshalIndent(serializedData, "", "  ")
// 	}

// 	if err != nil {
// 		return nil, err
// 	}

// 	return result, nil
// }

// // serializeUnsupported replaces unsupported types like functions with string descriptors.
// func serializeUnsupported(object any, includePrivateFields bool) any {
// 	if v == nil {
// 		return nil
// 	}

// 	val := reflect.ValueOf(v)
// 	switch val.Kind() {
// 	case reflect.Pointer:
// 		// Dereference the pointer
// 		if val.IsNil() {
// 			return nil
// 		}
// 		// Handle pointers by marking them with a * prefix
// 		result := serializeUnsupported(val.Elem().Interface(), includePrivateFields)
// 		// If it's a pointer, prepend "*" to the key
// 		if reflect.TypeOf(result).Kind() == reflect.Map {
// 			m := result.(map[string]any)
// 			for key := range m {
// 				m["*"+key] = m[key] // Add '*' prefix to each key
// 				delete(m, key)      // Remove the old key
// 			}
// 		}
// 		return result
// 	case reflect.Struct:
// 		// Handle structs
// 		m := make(map[string]any)
// 		typ := reflect.TypeOf(v)
// 		for i := 0; i < val.NumField(); i++ {
// 			field := val.Field(i)
// 			fieldType := typ.Field(i)
// 			if !field.CanInterface() {
// 				// optional
// 				if includePrivateFields {
// 					m[fieldType.Name] = "[private]"
// 				}
// 				continue // Skip unexported fields
// 			}
// 			m[fieldType.Name] = serializeUnsupported(field.Interface(), includePrivateFields)
// 		}
// 		return m
// 	case reflect.Slice:
// 		// Handle slices
// 		result := make([]any, val.Len())
// 		for i := 0; i < val.Len(); i++ {
// 			result[i] = serializeUnsupported(val.Index(i).Interface(), includePrivateFields)
// 		}
// 		return result
// 	case reflect.Map:
// 		// Handle maps
// 		result := make(map[string]any)
// 		for _, key := range val.MapKeys() {
// 			result[fmt.Sprintf("%v", key.Interface())] = serializeUnsupported(val.MapIndex(key).Interface(), includePrivateFields)
// 		}
// 		return result
// 	default:
// 		if val.Kind() == reflect.Func {
// 			// Automatically generate a function descriptor
// 			funcType := val.Type()
// 			// funcName := funcType.Name()

// 			// Get parameter types of the function (if any)
// 			numParams := funcType.NumIn()
// 			paramTypes := make([]string, numParams)
// 			for i := 0; i < numParams; i++ {
// 				paramTypes[i] = funcType.In(i).String()
// 			}

// 			// Get return type if available
// 			var returnType string
// 			if funcType.NumOut() > 0 {
// 				returnType = funcType.Out(0).String()
// 			}

// 			// Create a descriptor string that includes the function name, parameter types, and return type if any
// 			if returnType != "" {
// 				return fmt.Sprintf("func(%s) %s", joinTypes(paramTypes), returnType)
// 			} else {
// 				return fmt.Sprintf("func(%s)", joinTypes(paramTypes))
// 			}
// 		}
// 		// Return the value directly for supported types
// 		return v
// 	}
// }

// // joinTypes is a helper function to join type strings with commas.
// func joinTypes(types []string) string {
// 	if len(types) == 0 {
// 		return ""
// 	}
// 	return fmt.Sprintf("%s", types)
// }

// package pprint

// import (
// 	"encoding/json"
// 	"fmt"
// 	"reflect"
// )

// // CustomMarshal replaces unsupported types with string descriptors and controls HTML escaping.
// func CustomMarshal(object any, escapeHTML bool, includePrivateFields bool) ([]byte, error) {
// 	// Marshal data with custom serialization
// 	serializedData := serializeUnsupported(v, includePrivateFields)

// 	// Marshal data with indentation and optional HTML escaping
// 	var result []byte
// 	var err error
// 	if escapeHTML {
// 		// If escapeHTML is true, use standard MarshalIndent
// 		result, err = json.MarshalIndent(serializedData, "", "  ")
// 	} else {
// 		// If escapeHTML is false, use encoder with SetEscapeHTML(false)
// 		encoder := json.NewEncoder(nil)
// 		encoder.SetEscapeHTML(false)
// 		result, err = json.MarshalIndent(serializedData, "", "  ")
// 	}

// 	if err != nil {
// 		return nil, err
// 	}

// 	return result, nil
// }

// // serializeUnsupported replaces unsupported types like functions with string descriptors.
// func serializeUnsupported(object any, includePrivateFields bool) any {
// 	if v == nil {
// 		return nil
// 	}

// 	val := reflect.ValueOf(v)
// 	switch val.Kind() {
// 	case reflect.Pointer:
// 		// Dereference the pointer
// 		if val.IsNil() {
// 			return nil
// 		}
// 		return serializeUnsupported(val.Elem().Interface(), includePrivateFields)
// 	case reflect.Struct:
// 		// Handle structs
// 		m := make(map[string]any)
// 		typ := reflect.TypeOf(v)
// 		for i := 0; i < val.NumField(); i++ {
// 			field := val.Field(i)
// 			fieldType := typ.Field(i)
// 			if !field.CanInterface() {
// 				// optional
// 				if includePrivateFields {
// 					m[fieldType.Name] = "[private]"
// 				}
// 				continue // Skip unexported fields
// 			}
// 			m[fieldType.Name] = serializeUnsupported(field.Interface(), includePrivateFields)
// 		}
// 		return m
// 	case reflect.Slice:
// 		// Handle slices
// 		result := make([]any, val.Len())
// 		for i := 0; i < val.Len(); i++ {
// 			result[i] = serializeUnsupported(val.Index(i).Interface(), includePrivateFields)
// 		}
// 		return result
// 	case reflect.Map:
// 		// Handle maps
// 		result := make(map[string]any)
// 		for _, key := range val.MapKeys() {
// 			result[fmt.Sprintf("%v", key.Interface())] = serializeUnsupported(val.MapIndex(key).Interface(), includePrivateFields)
// 		}
// 		return result
// 	default:
// 		if val.Kind() == reflect.Func {
// 			// Automatically generate a function descriptor
// 			funcType := val.Type()
// 			// funcName := funcType.Name()

// 			// Get parameter types of the function (if any)
// 			numParams := funcType.NumIn()
// 			paramTypes := make([]string, numParams)
// 			for i := 0; i < numParams; i++ {
// 				paramTypes[i] = funcType.In(i).String()
// 			}

// 			// Get return type if available
// 			var returnType string
// 			if funcType.NumOut() > 0 {
// 				returnType = funcType.Out(0).String()
// 			}

// 			// Create a descriptor string that includes the function name, parameter types, and return type if any
// 			if returnType != "" {
// 				return fmt.Sprintf("func(%s) %s", joinTypes(paramTypes), returnType)
// 			} else {
// 				return fmt.Sprintf("func(%s)", joinTypes(paramTypes))
// 			}
// 		}
// 		// Return the value directly for supported types
// 		return v
// 	}
// }

// // joinTypes is a helper function to join type strings with commas.
// func joinTypes(types []string) string {
// 	if len(types) == 0 {
// 		return ""
// 	}
// 	return fmt.Sprintf("%s", types)
// }
