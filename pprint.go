package pprint

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unsafe"
)

func PPrint(
	object any,
	stream io.Writer,
	indent, width, depth int,
	compact, sortDicts, underscoreNumbers bool,
) {

	printer, error := NewPrettyPrinter(stream, indent, width, depth, compact, sortDicts, underscoreNumbers)
	if error != nil {
		panic(error)
	}
	printer.PPrint(object)
}

func PFormat(
	object any,
	stream io.Writer,
	indent, width, depth int,
	compact, sortDicts, underscoreNumbers bool,
) string {
	printer, error := NewPrettyPrinter(stream, indent, width, depth, compact, sortDicts, underscoreNumbers)
	if error != nil {
		panic(error)
	}
	return printer.PFormat(object)
}

func PP(
	object any,
	stream io.Writer,
	indent, width, depth int,
	compact, sortDicts, underscoreNumbers bool,
) {
	PPrint(object, stream, indent, width, depth, compact, sortDicts, underscoreNumbers)
}

func SafeRepr(object any) any {
	str, _, _ := PrettyPrinter{}.safeRepr(object, map[uintptr]int{}, 0, 0)
	return str
}

func IsReadable(object any) any {
	_, readable, _ := PrettyPrinter{}.safeRepr(object, map[uintptr]int{}, 0, 0)
	return readable
}

func IsRecurcive(object any) any {
	_, _, recursive := PrettyPrinter{}.safeRepr(object, map[uintptr]int{}, 0, 0)
	return recursive
}

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

type PrettyPrinter struct {
	stream            io.Writer
	width             int
	depth             int
	indentPerLevel    int
	compact           bool
	sortDicts         bool
	underscoreNumbers bool

	recursive   bool
	readable    bool
	dispatchMap map[reflect.Type]pprinter
}

type PrettyPrinterInterface interface {
	PPrint(object any)                                                                                    // +
	PFormat(object any) string                                                                            // +
	IsRecursive(object any) bool                                                                          // +
	IsReadable(object any) bool                                                                           // +
	format(object any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int)       // +
	pprintStruct(object any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int) // pprint_dataclass
	pprintMap(object any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int)    // pprint_dict
	// pprintSlice(object []any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int) // pprint_list / pprint_tuple / print_set
	// pprintString(object []any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int) // pprint_string
	// pprintBytes(object []any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int) // pprint_bytes / pprint_bytearray
	// pprintMappingProxy(object []any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int) // pprint_bytes / _pprint_mappingproxy
	// pprintSimpleNameSpace(object []any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int) // pprint_bytes / _pprint_simplenamespace

	formatMapItems(object []map[any]any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int) // pprint_dict
	//	formatNameSpaceItems(items []map[any]any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int) // pprint_dict
	//	formatItems(items []map[any]any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int) // pprint_dict

	repr(object any, context map[uintptr]int, level int) string                                                                                  // +
	Format(object any, context map[uintptr]int, maxLevels, level int) (string, bool, bool)                                                       // +
	pprintDefaultMap(object map[any]any, defaultFactory func() any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int) // -
	// pprintCounter(object map[any]any, defaultFactory func() any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int)
	// pprintChainMap(object map[any]any, defaultFactory func() any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int)
	// pprintDeque(object map[any]any, defaultFactory func() any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int)
	// pprintUserMap(object map[any]any, defaultFactory func() any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int)
	// pprintUserSlice(object map[any]any, defaultFactory func() any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int)
	// pprintUserString(object map[any]any, defaultFactory func() any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int)
	safeRepr(object any, context map[uintptr]int, maxLevels, level int) (string, bool, bool) // +
}

type pprinter func(pp PrettyPrinter, object any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int)

var defaultDispatchMap = make(map[reflect.Type]pprinter)
var builtinScalars []any

func getType[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

func init() {
	defaultDispatchMap[reflect.TypeOf((*struct{})(nil)).Elem()] = PrettyPrinter.pprintStruct
	// defaultDispatchMap["[]any"] = PrettyPrinter.pprintSlice
	defaultDispatchMap[reflect.TypeOf(map[any]any{})] = PrettyPrinter.pprintMap

	builtinScalars = []any{
		getType[bool](),
		getType[int](),
		getType[int8](),
		getType[int16](),
		getType[int32](),
		getType[int64](),
		getType[float32](),
		getType[float64](),
		getType[complex64](),
		getType[complex128](),
		getType[string](),
		getType[rune](),
		getType[byte](),
	}
}

func NewPrettyPrinter(
	stream io.Writer,
	indent, width, depth int,
	compact, sortDicts, underscoreNumbers bool,
) (PrettyPrinterInterface, error) {
	// Validate parameters
	if indent < 0 {
		return nil, fmt.Errorf("indent must be >= 0")
	}
	if depth <= 0 {
		return nil, fmt.Errorf("depth must be > 0")
	}
	if width <= 0 {
		width = 80 // default value
	}

	// Use default stream if nil
	if stream == nil {
		stream = os.Stdout
	}

	// Return the initialized PrettyPrinter
	return PrettyPrinter{
		depth:             depth,
		indentPerLevel:    indent,
		width:             width,
		stream:            stream,
		compact:           compact,
		sortDicts:         sortDicts,
		underscoreNumbers: underscoreNumbers,
		dispatchMap:       defaultDispatchMap,
	}, nil
}

func (pp PrettyPrinter) PPrint(object any) {
	if pp.stream != nil {
		pp.format(object, pp.stream, 0, 0, nil, 0)
		io.WriteString(pp.stream, "\n") // Write newline
	}
}

func (pp PrettyPrinter) PFormat(object any) string {
	var sio bytes.Buffer
	pp.format(object, &sio, 0, 0, nil, 0) // Format the object into the buffer
	return sio.String()                   // Return the formatted content as string
}

func (pp PrettyPrinter) IsRecursive(object any) bool {
	_, _, recursive := pp.Format(object, nil, 0, 0)
	return recursive
}

func (pp PrettyPrinter) IsReadable(object any) bool {
	// Call the format method once and capture the returned values
	_, readable, _ := pp.Format(object, nil, 0, 0)
	return readable
}

func (pp PrettyPrinter) format(object any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int) {
	// Get the unique id of the object (using reflect to simulate id)
	objectId := getObjectId(object)
	if _, exists := context[objectId]; exists {
		io.WriteString(stream, recursion(object))
		// Recursion detected
		pp.recursive = true
		pp.readable = false
		return
	}

	// Get the string representation of the object
	rep := pp.repr(object, context, level)

	// Check if the representation exceeds the max width
	maxWidth := pp.width - indent - allowance

	if len(rep) > maxWidth {

		p, exists := pp.dispatchMap[reflect.TypeOf(object)]
		fmt.Println(p, exists)

		// // If object is wide, we need to dispatch for special handling (e.g., handling dataclasses)
		// // Here we check if the object is a struct (similar to dataclass in Python)
		// if reflect.TypeOf(object).Kind() == reflect.Struct {
		// 	// For structs (dataclass-like objects), we handle them specially
		// 	pp.pprintStruct(object, stream, indent, allowance, context, level+1)
		// 	return
		// }
	}

	// Write the normal representation to the stream
	_, _ = stream.Write([]byte(rep))
}

// formatMapItems formats each item of the dictionary (key-value pair)
func (pp PrettyPrinter) formatMapItems(items []map[any]any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int) {
	for _, item := range items {
		for k, v := range item {
			// Format each key-value pair by calling the _format method (simplified)
			// You would replace this with the actual formatting logic you want
			io.WriteString(stream, fmt.Sprintf("\n%s: %v", k, v))
		}
	}
}

// pprintMap formats the dictionary as a string
func (pp PrettyPrinter) pprintMap(object any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int) {
	_object := object.(map[any]any)
	io.WriteString(stream, "{")
	if pp.indentPerLevel > 1 {
		io.WriteString(stream, fmt.Sprintf("%*s", pp.indentPerLevel-1, " "))
	}
	length := len(_object)
	if length > 0 {
		// Sort dictionary items if required
		var items []map[any]any
		for k, v := range _object {
			items = append(items, map[any]any{k: v})
		}

		if pp.sortDicts {
			// Sorting items based on the keys
			sort.SliceStable(items, func(i, j int) bool {
				// Sort by the key (extract the key and compare)
				for k1 := range items[i] {
					for k2 := range items[j] {
						return fmt.Sprintf("%v", k1) < fmt.Sprintf("%v", k2)
					}
				}
				return false
			})
		}

		// Format the dictionary items
		pp.formatMapItems(items, stream, indent, allowance+1, context, level)
	}
	io.WriteString(stream, "\n}")
}

func copyMap(original map[uintptr]int) map[uintptr]int {
	// Create a new map to store the copied data
	copy := make(map[uintptr]int, len(original))

	// Copy each key-value pair
	for key, value := range original {
		copy[key] = value
	}

	return copy
}

// pprintStruct is a helper method to handle structs like dataclasses in Python
func (pp PrettyPrinter) pprintStruct(object any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int) {
	// This would handle special printing for struct-like objects (dataclass equivalent)
	// Here, we are just printing the struct as-is for simplicity
	_, _ = stream.Write([]byte(fmt.Sprintf("%+v", object)))
}

// repr simulates the Python's repr function that returns a string representation of the object.
func (pp PrettyPrinter) repr(object any, context map[uintptr]int, level int) string {
	repr, readable, recursive := pp.Format(object, copyMap(context), pp.depth, level)
	if !readable {
		pp.readable = false
	}
	if recursive {
		recursive = true
	}
	return repr
}

func (pp PrettyPrinter) Format(object any, context map[uintptr]int, maxLevels, level int) (string, bool, bool) {
	return pp.safeRepr(object, context, maxLevels, level)
}

func (pp PrettyPrinter) pprintDefaultMap(object map[any]any, defaultFactory func() any, stream io.Writer, indent, allowance int, context map[uintptr]int, level int) {

	// Handle empty dictionary
	if len(object) == 0 {
		io.WriteString(stream, fmt.Sprintf("%v", object))
		return
	}

	// // Get the "default_factory" equivalent
	// rdf := fmt.Sprintf("%v", defaultFactory)

	// Print the class name and default factory
	clsName := "map"           // Using "map" for Go's default type
	indent += len(clsName) + 1 // Adjust indentation for the class name

	// io.WriteString(stream, fmt.Sprintf("%s(%s,\n%s", clsName, rdf, strings.Repeat(" ", indent)))

	// Print the dictionary items
	pp.pprintMap(object, stream, indent, allowance+1, context, level)

	// Close the representation
	io.WriteString(stream, ")")
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

func getObjectId(object any) uintptr {
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

func idInContext(objectId uintptr, context map[uintptr]int) bool {
	_, exists := context[objectId]
	return exists
}

func (pp PrettyPrinter) safeRepr(object any, context map[uintptr]int, maxLevels, level int) (string, bool, bool) {

	// Get the type of the object
	typ := reflect.TypeOf(object)

	// Check if the object is one of the basic scalar types (e.g., int, float)
	for _, element := range builtinScalars {
		if element == typ {
			return fmt.Sprintf("%v", object), true, false
		}
	}

	// Handle integer types (int, int32, int64, etc.)
	if typ.Kind() == reflect.Int {
		if pp.underscoreNumbers {
			// Format with underscores for readability
			return formatWithUnderscores(object.(int)), true, false
		}
		return fmt.Sprintf("%d", object), true, false
	}

	// Handle maps (map[any]any)
	if typ.Kind() == reflect.Map {
		value := reflect.ValueOf(object)
		if value.Len() == 0 {
			return "{}", true, false
		}

		// Get the unique ID of the object
		objectId := getObjectId(object)

		// Recursion limit handling
		if maxLevels > 0 && level >= maxLevels {
			return "{...}", false, idInContext(objectId, context)
		}

		// Prevent infinite recursion
		if idInContext(objectId, context) {
			return "{...}", false, true
		}

		// Track this object in the context
		context[objectId] = 1

		readable := true
		recursive := false
		components := []string{}

		// Increment recursion level
		level += 1

		// Get map keys and sort them if necessary
		keys := make([]any, 0, value.Len())
		for _, k := range value.MapKeys() {
			keys = append(keys, k.Interface()) // Convert reflect.Value to actual value
		}

		if pp.sortDicts {
			sort.Slice(keys, func(i, j int) bool {
				return fmt.Sprintf("%v", keys[i]) < fmt.Sprintf("%v", keys[j])
			})
		}

		// Iterate over sorted keys and process key-value pairs
		for _, k := range keys {
			kRepr, kReadable, kRecur := pp.Format(k, context, maxLevels, level)
			vRepr, vReadable, vRecur := pp.Format(value.MapIndex(reflect.ValueOf(k)).Interface(), context, maxLevels, level)
			components = append(components, fmt.Sprintf("%s: %s", kRepr, vRepr))
			readable = readable && kReadable && vReadable
			if kRecur || vRecur {
				recursive = true
			}
		}

		// Cleanup context after processing this object
		delete(context, objectId)

		// Return the formatted map representation
		return fmt.Sprintf("{%s}", strings.Join(components, ", ")), readable, recursive
	}

	// Handle slices (which corresponds to Python's list and tuple types)
	if typ.Kind() == reflect.Slice {
		value := reflect.ValueOf(object)
		if value.Len() == 0 {
			return "[]", true, false // Empty slice
		}

		// Determine format based on length of slice
		format := "[%s]"
		if value.Len() == 1 {
			format = "(%s,)" // Single-element tuple format
		}

		objectId := getObjectId(object)

		// Recursion limit handling
		if maxLevels > 0 && level >= maxLevels {
			return fmt.Sprintf(format, "..."), false, idInContext(objectId, context)
		}

		// Prevent infinite recursion
		if idInContext(objectId, context) {
			return recursion(object), false, true
		}

		// Track object in the context to handle recursion
		context[objectId] = 1

		readable := true
		recursive := false
		components := []string{}
		level += 1

		// Process each element in the slice
		for i := 0; i < value.Len(); i++ {
			elem := value.Index(i).Interface()
			elemRepr, elemReadable, elemRecur := pp.Format(elem, context, maxLevels, level)
			components = append(components, elemRepr)

			// Update readability and recursion flags
			if !elemReadable {
				readable = false
			}
			if elemRecur {
				recursive = true
			}
		}

		// Clean up context after processing
		delete(context, objectId)

		// Return the formatted string for the slice
		return fmt.Sprintf(format, strings.Join(components, ", ")), readable, recursive
	}

	// if typ.Kind() == reflect.Struct {
	// 	value := reflect.ValueOf(object)
	// 	readable := true
	// 	recursive := false
	// 	components := []string{}
	// 	level += 1
	// 	for i := 0; i < value.NumField(); i++ {
	// 		field := value.Field(i)
	// 		fieldName := typ.Field(i).Name
	// 		fRepr, fReadable, fRecur := pp.Format(field.Interface(), context, maxLevels, level)
	// 		components = append(components, fmt.Sprintf("%s: %s", fieldName, fRepr))
	// 		readable = readable && fReadable
	// 		recursive = recursive || fRecur
	// 	}
	// 	return fmt.Sprintf("{%s}", strings.Join(components, ", ")), readable, recursive
	// }

	rep := fmt.Sprintf("%v", object)

	return rep, true, true
}

func recursion(object interface{}) string {
	// Get the type of the object
	objectType := reflect.TypeOf(object).Name()

	// Get the memory address (simulating id in Python)
	// Use reflect.ValueOf to get the value, then convert to unsafe.Pointer
	objectValue := reflect.ValueOf(object)
	objectID := fmt.Sprintf("%v", unsafe.Pointer(objectValue.Pointer()))

	// Return the formatted string similar to Python's recursion function
	return fmt.Sprintf("<Recursion on %s with id=%s>", objectType, objectID)
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
