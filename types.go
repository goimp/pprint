package pprint

import (
	"io"
	"reflect"
)

type Context map[uintptr]int
type DispatchMap map[reflect.Kind]pprinter

type pprinter func(pp PrettyPrinter, object any, stream io.Writer, indent, allowance int, context Context, level int)

type MappingItem struct {
	Key   any
	Entry any
}

type StructField struct {
	Name  string
	Entry any
}

func (ctx Context) Contains(objectId uintptr) bool {
	_, exists := ctx[objectId]
	return exists
}

type InaccessibleField struct {
	Name   string
	Reason string
}

func (i InaccessibleField) String() string {
	return "<InaccessibleField>" // Custom string representation when printed
}

// func reprInaccessible() string {
// 	// return fmt.Sprintf("%v", reflect.TypeOf(InaccessibleField{}))
// 	return fmt.Sprintf("%v", InaccessibleField{})
// }
