package pprint

import (
	"io"
	"reflect"
)

var defaultDispatchMap = make(DispatchMap)
var builtinScalars []any

func PPrint(
	object any,
	stream io.Writer,
	indent, width, depth int,
	compact, sortMaps, underscoreNumbers bool,
) {

	printer, error := NewPrettyPrinter(stream, indent, width, depth, compact, sortMaps, underscoreNumbers)
	if error != nil {
		panic(error)
	}
	printer.PPrint(object)
}

func PFormat(
	object any,
	stream io.Writer,
	indent, width, depth int,
	compact, sortMaps, underscoreNumbers bool,
) string {
	printer, error := NewPrettyPrinter(stream, indent, width, depth, compact, sortMaps, underscoreNumbers)
	if error != nil {
		panic(error)
	}
	return printer.PFormat(object)
}

func PP(
	object any,
	stream io.Writer,
	indent, width, depth int,
	compact, sortMaps, underscoreNumbers bool,
) {
	PPrint(object, stream, indent, width, depth, compact, sortMaps, underscoreNumbers)
}

func SafeRepr(object any) any {
	str, _, _ := PrettyPrinter{}.safeRepr(object, Context{}, 0, 0)
	return str
}

func IsReadable(object any) any {
	_, readable, _ := PrettyPrinter{}.safeRepr(object, Context{}, 0, 0)
	return readable
}

func IsRecurcive(object any) any {
	_, _, recursive := PrettyPrinter{}.safeRepr(object, Context{}, 0, 0)
	return recursive
}

func init() {
	defaultDispatchMap[reflect.Map] = PrettyPrinter.pprintMap
	defaultDispatchMap[reflect.Slice] = PrettyPrinter.pprintSlice
	defaultDispatchMap[reflect.Struct] = PrettyPrinter.pprintStruct

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
