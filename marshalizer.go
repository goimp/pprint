// -------------------------------------------------------------------------------------------------------

package pprint

import (
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

type Serializer func(val reflect.Value, mr Marshalizer) any

type MarshalizerContext map[uintptr]string
type KindSerializerMap map[reflect.Kind]Serializer
type TypeSerializerMap map[reflect.Type]Serializer

type MarshalizerInterface interface {
	Serialize(object any) ([]byte, error)
	AddKind(kind reflect.Kind, serializer Serializer)
	RemoveKind(kind reflect.Kind)
}

type SerializersRegistry struct {
	kindSerializers KindSerializerMap
	typeSerializers TypeSerializerMap
}

type Marshalizer struct {
	context              MarshalizerContext
	escapeHTML           bool
	includePrivateFields bool
	registry             SerializersRegistry
}

func NewMarshalizer(includePrivateFields bool, escapeHTML bool) MarshalizerInterface {

	// if registry == nil {
	registry := SerializersRegistry{
		kindSerializers: make(KindSerializerMap),
		typeSerializers: make(TypeSerializerMap),
	}
	registry.kindSerializers[reflect.Slice] = SerializeSlice
	registry.kindSerializers[reflect.Map] = SerializeMap
	registry.kindSerializers[reflect.Struct] = SerializeStruct
	registry.kindSerializers[reflect.Func] = SerializeFuncSignature
	registry.kindSerializers[reflect.Pointer] = SerializePointer

	// }

	mr := &Marshalizer{
		context:              make(MarshalizerContext),
		escapeHTML:           escapeHTML,
		includePrivateFields: includePrivateFields,
		registry:             registry,
	}

	return mr
}

func (mr Marshalizer) Serialize(object any) ([]byte, error) {
	// Marshal data with custom serialization
	serializedData := serialize(object, mr)

	// Marshal data with indentation and optional HTML escaping
	var result []byte
	var err error
	if mr.escapeHTML {
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

func (mr Marshalizer) AddKind(kind reflect.Kind, serializer Serializer) {
	if _, exists := mr.registry.kindSerializers[kind]; exists {
		panic(fmt.Sprintf("kind %s already registered", kind))
	}
	mr.registry.kindSerializers[kind] = serializer
}

func (mr Marshalizer) RemoveKind(kind reflect.Kind) {
	if _, exists := mr.registry.kindSerializers[kind]; !exists {
		panic(fmt.Sprintf("kind %s not in registry", kind))
	}
	delete(mr.registry.kindSerializers, kind)
}

// serialize replaces unsupported types like functions with string descriptors.
func serialize(object any, mr Marshalizer) any {
	if object == nil {
		return nil
	}

	val := reflect.ValueOf(object)

	if serializer, exists := mr.registry.typeSerializers[val.Type()]; exists {
		return serializer(val, mr)
	}

	if serializer, exists := mr.registry.kindSerializers[val.Kind()]; exists {
		return serializer(val, mr)
	}

	return object
}

func SerializePointer(val reflect.Value, mr Marshalizer) any {
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
	value := serialize(val.Elem().Interface(), mr)

	// If the dereferenced value is a map (structure-like), add pointer data
	if reflect.ValueOf(value).Kind() == reflect.Map {
		// Add pointer metadata to the map
		value.(map[string]any)["*"] = ptrData
		return value
	}
	// If it's not a map, just return the value as is
	return value
}

func SerializeStruct(val reflect.Value, mr Marshalizer) any {
	// Handle structs
	m := make(map[string]any)
	typ := val.Type()
	// typ := reflect.TypeOf(object)
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		if !field.CanInterface() {
			// optional
			if mr.includePrivateFields {
				m[fieldType.Name] = "[private]"
			}
			continue // Skip unexported fields
		}
		m[fieldType.Name] = serialize(field.Interface(), mr)
	}
	return m
}

func SerializeSlice(val reflect.Value, mr Marshalizer) any {
	// Handle slices
	result := make([]any, val.Len())
	for i := 0; i < val.Len(); i++ {
		result[i] = serialize(val.Index(i).Interface(), mr)
	}
	return result
}

func SerializeMap(val reflect.Value, mr Marshalizer) any {
	// Handle maps
	result := make(map[string]any)
	for _, key := range val.MapKeys() {
		result[fmt.Sprintf("%v", key.Interface())] = serialize(val.MapIndex(key).Interface(), mr)
	}
	return result
}

func SerializeFuncSignature(val reflect.Value, mr Marshalizer) any {
	// Automatically generate a function descriptor
	funcType := val.Type()

	// Get parameter types
	params := []string{}
	for i := 0; i < funcType.NumIn(); i++ {
		if typeString := funcType.In(i).String(); len(typeString) > 0 {
			params = append(params, removeSpaces(typeString))
		}
	}

	// Get return types
	results := []string{}
	for i := 0; i < funcType.NumOut(); i++ {
		if typeString := funcType.Out(i).String(); len(typeString) > 0 {
			results = append(results, removeSpaces(typeString))
		}
	}

	// Construct the signature
	paramList := strings.Join(params, ", ")
	resultList := strings.Join(results, ", ")

	// resolve func name
	funcName := runtime.FuncForPC(val.Pointer()).Name()

	signature := ""

	if len(funcName) > 0 {
		signature += cleanFuncName(funcName) + " "
	}

	signature += fmt.Sprintf("func(%s)", paramList)

	if len(results) == 1 {
		signature += fmt.Sprintf(" %s", resultList)
	} else if len(results) > 1 {
		signature += fmt.Sprintf(" (%s)", resultList)
	}

	return strings.TrimSpace(signature)
}

// Function to clean up spaces between type names
func removeSpaces(typeStr string) string {
	return strings.ReplaceAll(typeStr, " ", "")
}

func cleanFuncName(fullName string) string {
	// Split the full name by "." and return the last part
	parts := strings.Split(fullName, "/")
	return parts[len(parts)-1]
}
