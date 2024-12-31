package pprint

import (
	"fmt"
	"reflect"
)

type KindSerializerMap map[reflect.Kind]Serializer
type TypeSerializerMap map[reflect.Type]Serializer
type KnownInterface map[reflect.Type]int

type SerializerRegistryInterface interface {
	AddKind(kind reflect.Kind, serializer Serializer)
	AddType(typ reflect.Type, serializer Serializer)
	AddKnownInterface(typ reflect.Type)
	RemoveKind(kind reflect.Kind)
	RemoveType(typ reflect.Type)
	RemoveKnownInterface(typ reflect.Type)
}

type SerializersRegistry struct {
	kindSerializers KindSerializerMap
	typeSerializers TypeSerializerMap
	knownInterfaces KnownInterface
}

func (sr SerializersRegistry) AddKind(kind reflect.Kind, serializer Serializer) {
	if _, exists := sr.kindSerializers[kind]; exists {
		panic(fmt.Sprintf("kind %s already registered", kind))
	}
	sr.kindSerializers[kind] = serializer
}

func (sr SerializersRegistry) RemoveKind(kind reflect.Kind) {
	if _, exists := sr.kindSerializers[kind]; !exists {
		panic(fmt.Sprintf("kind %s not in registry", kind))
	}
	delete(sr.kindSerializers, kind)
}

func (sr SerializersRegistry) AddType(typ reflect.Type, serializer Serializer) {
	if _, exists := sr.typeSerializers[typ]; exists {
		panic(fmt.Sprintf("type %s already registered", typ))
	}
	sr.typeSerializers[typ] = serializer
}

func (sr SerializersRegistry) RemoveType(typ reflect.Type) {
	if _, exists := sr.typeSerializers[typ]; !exists {
		panic(fmt.Sprintf("type %s not in registry", typ))
	}
	delete(sr.typeSerializers, typ)
}

func (sr SerializersRegistry) AddKnownInterface(typ reflect.Type) {
	if _, exists := sr.knownInterfaces[typ]; exists {
		panic(fmt.Sprintf("interface %s already registered", typ))
	}
	sr.knownInterfaces[typ] = 1
}

func (sr SerializersRegistry) RemoveKnownInterface(typ reflect.Type) {
	if _, exists := sr.knownInterfaces[typ]; !exists {
		panic(fmt.Sprintf("interface %s not in registry", typ))
	}
	delete(sr.knownInterfaces, typ)
}
