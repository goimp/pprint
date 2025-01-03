package pprint

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

type PrettyPrinter struct {
	stream            io.Writer
	width             int
	depth             int
	indentPerLevel    int
	compact           bool
	sortMaps          bool
	underscoreNumbers bool

	recursive   bool
	readable    bool
	dispatchMap DispatchMap
}

type PrettyPrinterInterface interface {
	PPrint(object any)                                                                            // +
	PFormat(object any) string                                                                    // +
	IsRecursive(object any) bool                                                                  // +
	IsReadable(object any) bool                                                                   // +
	format(object any, stream io.Writer, indent, allowance int, context Context, level int)       // +
	pprintMap(object any, stream io.Writer, indent, allowance int, context Context, level int)    // +
	pprintSlice(object any, stream io.Writer, indent, allowance int, context Context, level int)  // +
	pprintStruct(object any, stream io.Writer, indent, allowance int, context Context, level int) // +
	pprintString(object any, stream io.Writer, indent, allowance int, context Context, level int) // +
	pprintBytes(object any, stream io.Writer, indent, allowance int, context Context, level int)  // +

	formatMapItems(items []MappingItem, stream io.Writer, indent, allowance int, context Context, level int)    // +
	formatItems(object []any, stream io.Writer, indent, allowance int, context Context, level int)              // +
	formatStructItems(items []StructField, stream io.Writer, indent, allowance int, context Context, level int) // +

	repr(object any, context Context, level int) string                              // +
	Format(object any, context Context, maxLevels, level int) (string, bool, bool)   // +
	safeRepr(object any, context Context, maxLevels, level int) (string, bool, bool) // +
}

func NewPrettyPrinter(
	stream io.Writer,
	indent, width, depth int,
	compact, sortMaps, underscoreNumbers bool,
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
		sortMaps:          sortMaps,
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

func (pp PrettyPrinter) format(object any, stream io.Writer, indent, allowance int, context Context, level int) {
	// Get the unique id of the object (using reflect to simulate id)
	objectId := id(object)

	if context == nil {
		context = make(Context)
	}

	if _, exists := context[objectId]; exists {
		io.WriteString(stream, recursion(object))
		// Recursion detected
		pp.recursive = true
		pp.readable = false
		return
	}

	rep := pp.repr(object, context, level)
	// Check if the representation exceeds the max width
	maxWidth := pp.width - indent - allowance

	if len(rep) > maxWidth {
		if object == nil {
			io.WriteString(stream, rep)
			return
		}
		typ := reflect.TypeOf(object)
		p, exists := pp.dispatchMap[typ.Kind()]
		if exists {
			context[objectId] = 1
			p(pp, object, stream, indent, allowance, context, level+1)
			delete(context, objectId)
			return
		}
	}

	// Write the normal representation to the stream
	io.WriteString(stream, rep)
}

func (pp PrettyPrinter) pprintPointer(object any, stream io.Writer, indent, allowance int, context Context, level int) {
	// objectId := id(object)
	// typ := reflect.TypeOf(object)
	value := reflect.ValueOf(object)
	if value.Kind() == reflect.Pointer {
		pointerPrefix := fmt.Sprintf("(%T=%p)&", object, object)
		io.WriteString(stream, pointerPrefix)
		intf := reflect.Indirect(value).Interface()
		// indent += len(pointerPrefix)
		indent += 1
		pp.format(intf, stream, indent, allowance, context, level)
	}
}

func (pp PrettyPrinter) pprintMap(object any, stream io.Writer, indent, allowance int, context Context, level int) {
	io.WriteString(stream, "{")
	if pp.indentPerLevel > 1 {
		io.WriteString(stream, strings.Repeat(" ", pp.indentPerLevel-1))
	}

	if mapping, ok := object.(map[any]any); ok {
		length := len(mapping)
		if length > 0 {

			var items []MappingItem

			for key, value := range mapping {
				items = append(items, MappingItem{Key: key, Entry: value})
			}

			// FIXME:
			// if pp.sortMaps {
			// 	items = sort.Slice(mapping, safeKey)
			// }
			pp.formatMapItems(items, stream, indent, allowance+1, context, level)
		}
	}

	io.WriteString(stream, "}")
}

func (pp PrettyPrinter) formatMapItems(items []MappingItem, stream io.Writer, indent, allowance int, context Context, level int) {
	indent += pp.indentPerLevel
	delimnl := ",\n" + strings.Repeat(" ", indent)
	lastIndex := len(items) - 1
	for i, item := range items {
		last := i == lastIndex
		rep := pp.repr(item.Key, context, level)
		io.WriteString(stream, rep)
		io.WriteString(stream, ": ")

		pp.format(item.Entry, stream, indent+len(rep)+2, allowance, context, level)
		if !last {
			io.WriteString(stream, delimnl)
		}
	}
}

func (pp PrettyPrinter) pprintSlice(object any, stream io.Writer, indent, allowance int, context Context, level int) {
	// Wrap []byte
	if slice, ok := object.([]byte); ok {
		pp.pprintBytes(slice, stream, indent, allowance, context, level)
		return
	}

	// Write the opening bracket for the slice
	io.WriteString(stream, "[")
	// Ensure the object is a slice, then call formatItems to handle the items

	if slice, ok := object.([]any); ok {
		pp.formatItems(slice, stream, indent, allowance+1, context, level)
	}
	// Write the closing bracket for the slice
	io.WriteString(stream, "]")
}

func (pp PrettyPrinter) formatItems(items []any, stream io.Writer, indent, allowance int, context Context, level int) {
	// Increase indent for the next level
	indent += pp.indentPerLevel
	if pp.indentPerLevel > 1 {
		io.WriteString(stream, fmt.Sprintf("%*s", pp.indentPerLevel-1, ""))
	}

	delimnl := ",\n" + fmt.Sprintf("%*s", indent, "")
	delim := ""
	width := pp.width - indent + 1
	maxWidth := width

	for i, ent := range items {
		// Check if it's the last item
		last := i == len(items)-1
		if last {
			maxWidth -= allowance
			width -= allowance
		}

		if pp.compact {
			rep := pp.repr(ent, context, level)
			w := len(rep) + 2
			if width < w {
				width = maxWidth
				if delim != "" {
					delim = delimnl
				}
			}
			if width >= w {
				width -= w
				io.WriteString(stream, delim)
				delim = ", "
				io.WriteString(stream, rep)
				continue
			}
		}

		io.WriteString(stream, delim)
		delim = delimnl
		pp.format(ent, stream, indent, allowance, context, level)
	}
}

func (pp PrettyPrinter) pprintStruct(object any, stream io.Writer, indent, allowance int, context Context, level int) {

	value := reflect.ValueOf(object)

	if value.Kind() == reflect.Struct {
		// Get the name of the struct
		typ := reflect.TypeOf(object)
		structName := typ.Name()

		var items []StructField

		// Now you can process the fields of the struct
		for i := 0; i < value.NumField(); i++ {
			field := value.Field(i)
			fieldName := typ.Field(i).Name
			// fmt.Println("isValid", field.IsValid(), "canSet", field.CanSet(), "canInterface", field.CanInterface())
			// if field.IsValid() && field.CanSet() {
			if field.IsValid() && field.CanInterface() {
				items = append(items, StructField{
					Name:  fieldName,
					Entry: field.Interface(), // Access the field as interface{}
				})
			} else {
				items = append(items, StructField{
					Name:  fieldName,
					Entry: InaccessibleField{},
				})
			}
		}

		io.WriteString(stream, structName+"(\n")
		pp.formatStructItems(items, stream, indent+1, allowance, context, level)
		io.WriteString(stream, "\n"+strings.Repeat(" ", indent)+")")
	}
}

func (pp PrettyPrinter) formatStructItems(items []StructField, stream io.Writer, indent, allowance int, context Context, level int) {
	delimnl := ",\n"
	delim := strings.Repeat(" ", indent+1)
	lastIndex := len(items) - 1
	for i, item := range items {
		last := i == lastIndex

		io.WriteString(stream, delim)
		io.WriteString(stream, item.Name)
		io.WriteString(stream, "=")

		// if idInContext(id(item.Entry), context) {
		if context.Contains(id(item.Entry)) {
			io.WriteString(stream, "...")
		} else {
			pp.format(item.Entry, stream, indent+len(item.Name)+2, allowance, context, level)
		}

		if !last {
			io.WriteString(stream, delimnl)
		}
	}
}

func (pp PrettyPrinter) pprintString(object any, stream io.Writer, indent, allowance int, context Context, level int) {
	// io.WriteString()
	if str, ok := object.(string); ok {
		if len(str) == 0 {
			// Assuming repr is implemented to return the string's representation
			io.WriteString(stream, repr(object))
			return
		}
		chunks := []string{}
		lines := strings.Split(str, "\n")
		if level == 1 {
			indent += 1
			allowance += 1
		}
		maxWidth1 := pp.width - indent
		maxWidth := maxWidth1 // You don't need a pointer here, just use the value directly

		rep := ""
		for i, line := range lines {
			rep = repr(object) // Ensure repr is properly implemented in Go
			if i == len(lines)-1 {
				maxWidth1 -= allowance
			}
			if len(rep) <= maxWidth1 {
				chunks = append(chunks, rep)
			} else {
				re := regexp.MustCompile(`\S*\s*`)
				parts := re.FindAllString(line, -1)
				if len(parts) == 0 {
					fmt.Println("No parts found")
					return
				}
				if parts[len(parts)-1] != "" {
					fmt.Println("Last part should be empty")
					return
				}
				parts = parts[:len(parts)-1]
				maxWidth2 := maxWidth
				current := ""
				for j, part := range parts {
					candidate := current + part
					if j == len(parts)-1 && i == len(lines)-1 {
						maxWidth2 -= allowance
					}
					if len(repr(candidate)) > maxWidth2 {
						if len(current) > 0 {
							chunks = append(chunks, repr(current))
						}
						current = part
					} else {
						current = candidate
					}
				}
				if len(current) > 0 {
					chunks = append(chunks, repr(current))
				}
			}
		}
		if len(chunks) == 1 {
			io.WriteString(stream, rep)
			return
		}
		if level == 1 {
			io.WriteString(stream, "(")
		}
		for i, rep := range chunks {
			if i > 0 {
				io.WriteString(stream, "\n"+strings.Repeat(" ", indent))
			}
			io.WriteString(stream, rep)
		}
		if level == 1 {
			io.WriteString(stream, ")")
		}
	}
}

func (pp PrettyPrinter) pprintBytes(object any, stream io.Writer, indent, allowance int, context Context, level int) {
	if data, ok := object.([]byte); ok {
		if len(data) <= 4 {
			io.WriteString(stream, fmt.Sprintf("%x", data))
			return
		}

		parens := level == 1
		if parens {
			indent++
			allowance++
			io.WriteString(stream, "(")
		}

		delim := ""
		wrapped := wrapBytesRepr(data, pp.width-indent, allowance) // Assuming 80 as the default width.
		for _, line := range wrapped {
			io.WriteString(stream, delim)
			// io.WriteString(stream, strings.Repeat(" ", indent))
			io.WriteString(stream, line)
			if delim == "" {
				delim = "\n"
			}
		}

		if parens {
			io.WriteString(stream, ")")
		}
	}
}

// repr simulates the Python's repr function that returns a string representation of the object.
func (pp PrettyPrinter) repr(object any, context Context, level int) string {
	repr, readable, recursive := pp.Format(object, copyContext(context), pp.depth, level)
	if !readable {
		pp.readable = false
	}
	if recursive {
		recursive = true
	}
	return repr
}

func (pp PrettyPrinter) Format(object any, context Context, maxLevels, level int) (string, bool, bool) {
	return pp.safeRepr(object, context, maxLevels, level)
}

func (pp PrettyPrinter) safeRepr(object any, context Context, maxLevels, level int) (string, bool, bool) {

	if object == nil {
		return repr(object), false, false
	}

	if object, ok := object.(InaccessibleField); ok {
		return object.String(), false, false
	}

	// Get the type of the object
	typ := reflect.TypeOf(object)

	// Check if the object is one of the basic scalar types (e.g., int, float)
	for _, element := range builtinScalars {
		if element == typ {
			return repr(object), true, false
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
		objectId := id(object)

		// Recursion limit handling
		if maxLevels > 0 && level >= maxLevels {
			// return "{...}", false, idInContext(objectId, context)
			return "{...}", false, context.Contains(objectId)
		}

		// Prevent infinite recursion
		// if idInContext(objectId, context) {
		if context.Contains(objectId) {
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

		if pp.sortMaps {
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

		objectId := id(object)

		// Recursion limit handling
		if maxLevels > 0 && level >= maxLevels {
			// return fmt.Sprintf(format, "..."), false, idInContext(objectId, context)
			return fmt.Sprintf(format, "..."), false, context.Contains(objectId)
		}

		// Prevent infinite recursion
		// if idInContext(objectId, context) {
		if context.Contains(objectId) {
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

	rep := repr(object)

	return rep, true, false
}
