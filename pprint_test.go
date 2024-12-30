package pprint

import (
	"os"
	"testing"
)

func TestPPrint(t *testing.T) {
	i := 1
	PPrint(i, nil, 1, 80, 2, false, true, false)

	s := "sample string"
	PPrint(s, nil, 1, 80, 2, false, true, false)

	f := 0.123
	PPrint(f, nil, 1, 80, 2, false, true, false)

	l := []any{1, "sample text", true}
	PPrint(l, nil, 1, 80, 2, false, true, false)

	m := map[any]any{
		"1": 11,
		2:   22,
	}
	PPrint(m, nil, 1, 80, 2, false, true, false)

	nL := []any{
		1, "sample text", true,
		m, l,
	}

	PPrint(nL, nil, 1, 80, 2, false, true, false)

	nM := map[any]any{
		"1":     11,
		2:       22,
		"slice": l,
	}

	PPrint(nM, nil, 1, 80, 2, false, true, false)

	type P struct {
		f1 int
		f2 string
	}

	p := P{f1: 1, f2: "sample string"}
	PPrint(p, nil, 1, 80, 2, false, true, false)
}

func TestPPrintFormat(t *testing.T) {
	ppi, err := NewPrettyPrinter(
		os.Stdout, 1, 80, 2, false, true, false,
	)

	pp := ppi.(PrettyPrinter)

	if err != nil {
		t.Error(err)
	}

	// Test with a simple object
	testObject := "This is a test object with a long representation"
	pp.format(testObject, pp.stream, 2, 0, make(map[uintptr]int), 0)

	// Test with a struct (similar to a dataclass)
	testStruct := struct {
		Name  string
		Value int
	}{
		Name:  "Test Struct",
		Value: 42,
	}
	pp.format(testStruct, pp.stream, 2, 0, make(map[uintptr]int), 0)
}
