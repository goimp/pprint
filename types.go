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

// func idInContext(objectId uintptr, context Context) bool {
// 	_, exists := context[objectId]
// 	return exists
// }

func (ctx Context) Contains(objectId uintptr) bool {
	_, exists := ctx[objectId]
	return exists
}
