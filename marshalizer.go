// -------------------------------------------------------------------------------------------------------

package pprint

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
)

type Serializer func(val reflect.Value, mr Marshalizer) any

type MarshalizerContext map[uintptr]int

func (ctx MarshalizerContext) Contains(objectId uintptr) bool {
	_, exists := ctx[objectId]
	return exists
}

func (ctx MarshalizerContext) Set(objectId uintptr) {
	ctx[objectId] = 1
}

func (ctx MarshalizerContext) Del(objectId uintptr) {
	delete(ctx, objectId)
}

type MarshalizerInterface interface {
	Serialize(object any) ([]byte, error)
}

type Marshalizer struct {
	context              MarshalizerContext
	escapeHTML           bool
	includePrivateFields bool
	includeImplements    bool
	registry             SerializersRegistry
}

func NewMarshalizer(includePrivateFields bool, escapeHTML bool, emptyRegistry bool, includeImplements bool) MarshalizerInterface {

	registry := SerializersRegistry{
		kindSerializers: make(KindSerializerMap),
		typeSerializers: make(TypeSerializerMap),
		knownInterfaces: make(KnownInterface),
	}

	if !emptyRegistry {
		registry.AddKind(reflect.Slice, SerializeSlice)
		registry.AddKind(reflect.Map, SerializeMap)
		registry.AddKind(reflect.Struct, SerializeStruct)
		registry.AddKind(reflect.Func, SerializeFuncSignature)
		registry.AddKind(reflect.Pointer, SerializePointer)

		registry.AddKnownInterface(reflect.TypeOf((*fmt.Stringer)(nil)).Elem())
		registry.AddKnownInterface(reflect.TypeOf((*fmt.Scanner)(nil)).Elem())
		registry.AddKnownInterface(reflect.TypeOf((*fmt.Formatter)(nil)).Elem())
		registry.AddKnownInterface(reflect.TypeOf((*error)(nil)).Elem())
		registry.AddKnownInterface(reflect.TypeOf((*io.Reader)(nil)).Elem())
		registry.AddKnownInterface(reflect.TypeOf((*io.Writer)(nil)).Elem())
		registry.AddKnownInterface(reflect.TypeOf((*io.Closer)(nil)).Elem())
		registry.AddKnownInterface(reflect.TypeOf((*io.ReadWriter)(nil)).Elem())
		registry.AddKnownInterface(reflect.TypeOf((*io.ReadSeeker)(nil)).Elem())
		registry.AddKnownInterface(reflect.TypeOf((*io.Seeker)(nil)).Elem())
		registry.AddKnownInterface(reflect.TypeOf((*io.WriteSeeker)(nil)).Elem())
		registry.AddKnownInterface(reflect.TypeOf((*io.ReadWriteSeeker)(nil)).Elem())
		registry.AddKnownInterface(reflect.TypeOf((*io.ReadWriteCloser)(nil)).Elem())
		registry.AddKnownInterface(reflect.TypeOf((*io.WriterAt)(nil)).Elem())
		registry.AddKnownInterface(reflect.TypeOf((*io.ReaderAt)(nil)).Elem())
		registry.AddKnownInterface(reflect.TypeOf((*sync.Locker)(nil)).Elem())
		// registry.AddKnownInterface(reflect.TypeOf((*sync.Mutex)(nil)).Elem()) // non interface
		// registry.AddKnownInterface(reflect.TypeOf((*sync.RWMutex)(nil)).Elem()) // non interface
		// registry.AddKnownInterface(reflect.TypeOf((*sync.Atomic)(nil)).Elem()) // unimplemented ?
		// registry.AddKnownInterface(reflect.TypeOf((*sync.WaitGroup)(nil)).Elem()) // non interface
		registry.AddKnownInterface(reflect.TypeOf((*http.RoundTripper)(nil)).Elem())
		registry.AddKnownInterface(reflect.TypeOf((*http.Handler)(nil)).Elem())
		// registry.AddKnownInterface(reflect.TypeOf((*http.ServeHTTP)(nil)).Elem()) // unimplemented ?
		registry.AddKnownInterface(reflect.TypeOf((*context.Context)(nil)).Elem())
		// registry.AddKnownInterface(reflect.TypeOf((*context.CancelFunc)(nil)).Elem()) // non interface
		registry.AddKnownInterface(reflect.TypeOf((*sort.Interface)(nil)).Elem())
		registry.AddKnownInterface(reflect.TypeOf((*testing.TB)(nil)).Elem())
		// registry.AddKnownInterface(reflect.TypeOf((*sql.Scanner)(nil)).Elem()) // unimplemented ?
		// registry.AddKnownInterface(reflect.TypeOf((*sql.Valuer)(nil)).Elem()) // unimplemented ?
		// registry.AddKnownInterface(reflect.TypeOf((*strconv.NumError)(nil)).Elem()) // non interface
		// registry.AddKnownInterface(reflect.TypeOf((*os.File)(nil)).Elem()) // non interface
		registry.AddKnownInterface(reflect.TypeOf((*net.Conn)(nil)).Elem())

		registry.AddKnownInterface(reflect.TypeOf((*MarshalizerInterface)(nil)).Elem())
		registry.AddKnownInterface(reflect.TypeOf((*SerializerRegistryInterface)(nil)).Elem())
	}

	mr := &Marshalizer{
		context:              make(MarshalizerContext),
		escapeHTML:           escapeHTML,
		includePrivateFields: includePrivateFields,
		includeImplements:    includeImplements,
		registry:             registry,
	}

	return mr
}

func (mr Marshalizer) String() string {
	result, err := mr.Serialize(mr)
	if err != nil {
		panic("can't serialize Marshalizer itself")
	}
	return string(result)
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
	mr.registry.AddKind(kind, serializer)
}

func (mr Marshalizer) RemoveKind(kind reflect.Kind) {
	mr.registry.RemoveKind(kind)
}

func (mr Marshalizer) AddType(typ reflect.Type, serializer Serializer) {
	mr.registry.AddType(typ, serializer)
}

func (mr Marshalizer) RemoveType(typ reflect.Type) {
	mr.registry.RemoveType(typ)
}

func (mr Marshalizer) AddKnownInterface(typ reflect.Type) {
	mr.registry.AddKnownInterface(typ)
}

func (mr Marshalizer) RemoveKnownInterface(typ reflect.Type) {
	mr.registry.RemoveKnownInterface(typ)
}

// serialize replaces unsupported types like functions with string descriptors.
func serialize(object any, mr Marshalizer) any {
	if object == nil {
		return nil
	}

	objectId := id(object)
	if mr.context.Contains(objectId) {
		// return fmt.Sprintf("(%T=%p)[Recursion Exceeded]", object, object)
		return fmt.Sprintf("(%T=%p)[Recursion Exceeded]", object, object)
	}

	mr.context.Set(objectId)

	val := reflect.ValueOf(object)

	if serializer, exists := mr.registry.typeSerializers[val.Type()]; exists {
		r := serializer(val, mr)
		mr.context.Del(objectId)
		return r
	}

	if serializer, exists := mr.registry.kindSerializers[val.Kind()]; exists {
		r := serializer(val, mr)
		mr.context.Del(objectId)
		return r
	}

	mr.context.Del(objectId)

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
		"type":    fmt.Sprintf("%T", val.Interface()), // Pointer type
	}

	if mr.includeImplements {
		interfaces := GetImplementedInterfacesDescriptor(val, mr)
		ptrData["implements"] = interfaces
	}

	// Serialize the dereferenced value first
	value := serialize(val.Elem().Interface(), mr)

	// If the dereferenced value is a map (structure-like), add pointer data
	if reflect.ValueOf(value).Kind() == reflect.Map {
		// Add pointer metadata to the map
		value.(map[string]any)["*"] = ptrData
		return value
	} else {
		output := map[string]any{
			"*":      ptrData,
			"_value": value,
		}
		return output
	}
	// If it's not a map, just return the value as is
	// return value
}

// DiscoverInterfaces dynamically finds all interfaces implemented by a given struct.
func DiscoverInterfaces(structType reflect.Type, interfaces KnownInterface) []reflect.Type {
	implemented := []reflect.Type{}

	// Iterate over the known interfaces map
	for ifaceType := range interfaces {
		if structType.Implements(ifaceType) {
			implemented = append(implemented, ifaceType)
		}
	}

	return implemented
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
				m[fieldType.Name] = "[Private Field]"
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

	// resolve func name
	funcName := runtime.FuncForPC(val.Pointer()).Name()

	return joinFuncSignature(funcName, params, results)
}

func SerializeMethodSignature(method reflect.Method, mr Marshalizer) string {
	// Access the method's type
	methodType := method.Type
	methodName := method.Name
	// Get parameter types (excluding the receiver)
	params := []string{}
	for i := 1; i < methodType.NumIn(); i++ { // Skip the first parameter, which is the receiver
		paramType := methodType.In(i).String()
		params = append(params, removeSpaces(paramType))
	}

	// Get return types
	results := []string{}
	for i := 0; i < methodType.NumOut(); i++ {
		resultType := methodType.Out(i).String()
		results = append(results, removeSpaces(resultType))
	}

	return joinFuncSignature(methodName, params, results)
}

func GetImplementedInterfacesDescriptor(val reflect.Value, mr Marshalizer) map[string][]string {
	implementedInterfaces := DiscoverInterfaces(val.Type(), mr.registry.knownInterfaces)

	serializedInterfaces := map[string][]string{}

	if len(implementedInterfaces) > 0 {
		for _, intf := range implementedInterfaces {
			// if serializedInterface := GetInterfaceDescriptor(intf, mr); serializedInterface != nil {
			// 	serializedInterfaces[intf.String()] = serializedInterface
			// }
			serializedInterfaces[intf.String()] = GetInterfaceDescriptor(intf, mr)
		}
	}
	if len(serializedInterfaces) > 0 {
		return serializedInterfaces
	}
	return nil
}

func GetInterfaceDescriptor(typ reflect.Type, mr Marshalizer) []string {
	if typ.Kind() == reflect.Interface {
		// If the interface is not nil, inspect the methods implemented by it
		methods := []string{}

		// Loop through all methods of the interface
		numMethods := typ.NumMethod()
		for i := 0; i < numMethods; i++ {
			method := typ.Method(i)

			// Handle the method signature separately
			methodSignature := SerializeMethodSignature(method, mr)
			methods = append(methods, methodSignature)
		}
		return methods
	}
	return nil
}

func joinFuncSignature(name string, params, results []string) string {
	signature := ""

	paramList := strings.Join(params, ", ")
	resultList := strings.Join(results, ", ")

	if len(name) > 0 {
		signature += cleanFuncName(name) + " "
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
