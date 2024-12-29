package pprint

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
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
	str, _, _ := PrettyPrinter{}.safeRepr(object, map[any]any{}, 0, 0)
	return str
}

func IsReadable(object any) any {
	_, readable, _ := PrettyPrinter{}.safeRepr(object, map[any]any{}, 0, 0)
	return readable
}

func IsRecurcive(object any) any {
	_, _, recursive := PrettyPrinter{}.safeRepr(object, map[any]any{}, 0, 0)
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
	dispatchMap map[string]pprinter
}

type PrettyPrinterInterface interface {
	PPrint(object any)                                                                                // +
	PFormat(object any) string                                                                        // +
	IsRecursive(object any) bool                                                                      // +
	IsReadable(object any) bool                                                                       // +
	format(object any, stream io.Writer, indent, allowance int, context map[any]any, level int)       // +
	pprintStruct(object any, stream io.Writer, indent, allowance int, context map[any]any, level int) // pprint_dataclass
	pprintMap(object any, stream io.Writer, indent, allowance int, context map[any]any, level int)    // pprint_dict
	// pprintSlice(object []any, stream io.Writer, indent, allowance int, context map[any]any, level int) // pprint_list / pprint_tuple / print_set
	// pprintString(object []any, stream io.Writer, indent, allowance int, context map[any]any, level int) // pprint_string
	// pprintBytes(object []any, stream io.Writer, indent, allowance int, context map[any]any, level int) // pprint_bytes / pprint_bytearray
	// pprintMappingProxy(object []any, stream io.Writer, indent, allowance int, context map[any]any, level int) // pprint_bytes / _pprint_mappingproxy
	// pprintSimpleNameSpace(object []any, stream io.Writer, indent, allowance int, context map[any]any, level int) // pprint_bytes / _pprint_simplenamespace

	formatMapItems(object []map[any]any, stream io.Writer, indent, allowance int, context map[any]any, level int) // pprint_dict
	//	formatNameSpaceItems(items []map[any]any, stream io.Writer, indent, allowance int, context map[any]any, level int) // pprint_dict
	//	formatItems(items []map[any]any, stream io.Writer, indent, allowance int, context map[any]any, level int) // pprint_dict

	repr(object any, context map[any]any, level int) string                            // +
	Format(object any, context map[any]any, maxLevels, level int) (string, bool, bool) // +
	pprintDefaultMap(object map[any]any, defaultFactory func() any, stream io.Writer, indent, allowance int, context map[any]any, level int)
	// pprintCounter(object map[any]any, defaultFactory func() any, stream io.Writer, indent, allowance int, context map[any]any, level int)
	// pprintChainMap(object map[any]any, defaultFactory func() any, stream io.Writer, indent, allowance int, context map[any]any, level int)
	// pprintDeque(object map[any]any, defaultFactory func() any, stream io.Writer, indent, allowance int, context map[any]any, level int)
	// pprintUserMap(object map[any]any, defaultFactory func() any, stream io.Writer, indent, allowance int, context map[any]any, level int)
	// pprintUserSlice(object map[any]any, defaultFactory func() any, stream io.Writer, indent, allowance int, context map[any]any, level int)
	// pprintUserString(object map[any]any, defaultFactory func() any, stream io.Writer, indent, allowance int, context map[any]any, level int)
	safeRepr(object any, context map[any]any, maxLevels, level int) (string, bool, bool) // -
}

type pprinter func(pp PrettyPrinter, object any, stream io.Writer, indent, allowance int, context map[any]any, level int)

var defaultDispatchMap = make(map[string]pprinter)

func init() {
	defaultDispatchMap["any"] = PrettyPrinter.pprintStruct
	// defaultDispatchMap["[]any"] = PrettyPrinter.pprintSlice
	defaultDispatchMap["map[any]any"] = PrettyPrinter.pprintMap
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

func (pp PrettyPrinter) format(object any, stream io.Writer, indent, allowance int, context map[any]any, level int) {
	// Get the unique id of the object (using reflect to simulate id)
	objID := fmt.Sprintf("%p", object)
	if _, exists := context[objID]; exists {
		// Recursion detected
		pp.recursive = true
		pp.readable = false
		_, _ = stream.Write([]byte("Recursion detected"))
		return
	}

	// Get the string representation of the object
	rep := pp.repr(object, context, level)

	// Check if the representation exceeds the max width
	maxWidth := pp.width - indent - allowance
	if len(rep) > maxWidth {
		// If object is wide, we need to dispatch for special handling (e.g., handling dataclasses)
		// Here we check if the object is a struct (similar to dataclass in Python)
		if reflect.TypeOf(object).Kind() == reflect.Struct {
			// For structs (dataclass-like objects), we handle them specially
			pp.pprintStruct(object, stream, indent, allowance, context, level+1)
			return
		}
	}

	// Write the normal representation to the stream
	_, _ = stream.Write([]byte(rep))
}

// formatMapItems formats each item of the dictionary (key-value pair)
func (pp PrettyPrinter) formatMapItems(items []map[any]any, stream io.Writer, indent, allowance int, context map[any]any, level int) {
	for _, item := range items {
		for k, v := range item {
			// Format each key-value pair by calling the _format method (simplified)
			// You would replace this with the actual formatting logic you want
			io.WriteString(stream, fmt.Sprintf("\n%s: %v", k, v))
		}
	}
}

// pprintMap formats the dictionary as a string
func (pp PrettyPrinter) pprintMap(object any, stream io.Writer, indent, allowance int, context map[any]any, level int) {
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

func copyMap(original map[any]any) map[any]any {
	// Create a new map to store the copied data
	copy := make(map[any]any, len(original))

	// Copy each key-value pair
	for key, value := range original {
		copy[key] = value
	}

	return copy
}

// pprintStruct is a helper method to handle structs like dataclasses in Python
func (pp PrettyPrinter) pprintStruct(object any, stream io.Writer, indent, allowance int, context map[any]any, level int) {
	// This would handle special printing for struct-like objects (dataclass equivalent)
	// Here, we are just printing the struct as-is for simplicity
	_, _ = stream.Write([]byte(fmt.Sprintf("%+v", object)))
}

// repr simulates the Python's repr function that returns a string representation of the object.
func (pp PrettyPrinter) repr(object any, context map[any]any, level int) string {
	repr, readable, recursive := pp.Format(object, copyMap(context), pp.depth, level)
	if !readable {
		pp.readable = false
	}
	if recursive {
		recursive = true
	}
	return repr
}

func (pp PrettyPrinter) Format(object any, context map[any]any, maxLevels, level int) (string, bool, bool) {
	return pp.safeRepr(object, context, maxLevels, level)
}

func (pp PrettyPrinter) pprintDefaultMap(object map[any]any, defaultFactory func() any, stream io.Writer, indent, allowance int, context map[any]any, level int) {

	// Handle empty dictionary
	if len(object) == 0 {
		io.WriteString(stream, fmt.Sprintf("%v", object))
		return
	}

	// Get the "default_factory" equivalent
	rdf := fmt.Sprintf("%v", defaultFactory)

	// Print the class name and default factory
	clsName := "map"           // Using "map" for Go's default type
	indent += len(clsName) + 1 // Adjust indentation for the class name

	io.WriteString(stream, fmt.Sprintf("%s(%s,\n%s", clsName, rdf, strings.Repeat(" ", indent)))

	// Print the dictionary items
	pp.pprintMap(object, stream, indent, allowance+1, context, level)

	// Close the representation
	io.WriteString(stream, ")")
}

func (p PrettyPrinter) safeRepr(object any, context map[any]any, maxLevels, level int) (string, bool, bool) {
	return "nil", true, true
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
