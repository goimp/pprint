package pprint

import (
	"os"
	"testing"
)

type sampleType struct {
	F1      int
	F2      string
	F3      string
	F4      string
	F5      string
	F6      string
	F7      string
	F8      string
	F9      string
	F10     string
	F11     any
	private int
}

// Helper function to generate a SampleType instance
func createSampleType(baseString string, nM any) sampleType {
	return sampleType{
		F1:      1,
		F2:      baseString + "2",
		F3:      baseString + "3",
		F4:      baseString + "4",
		F5:      baseString + "5",
		F6:      baseString + "6",
		F7:      baseString + "7",
		F8:      baseString + "8",
		F9:      baseString + "9",
		F10:     baseString + "10",
		F11:     nM,
		private: 10,
	}
}

func TestPPrintScalars(t *testing.T) {
	i := 1
	ie := "1"
	if out := PFormat(i, nil, 1, 80, 2, false, true, false); out != ie {
		t.Errorf("expected 1, got %s", out)
	} else {
		PPrint(i, nil, 1, 80, 2, false, true, false)
	}
	s := "sample string"
	if out := PFormat(s, nil, 1, 80, 2, false, true, false); out != s {
		t.Errorf("expected 1, got %s", out)
	} else {
		PPrint(s, nil, 1, 80, 2, false, true, false)
	}
	f := 0.123
	fe := "0.123"
	if out := PFormat(f, nil, 1, 80, 2, false, true, false); out != fe {
		t.Errorf("expected 1, got %s", out)
	} else {
		PPrint(f, nil, 1, 80, 2, false, true, false)
	}
}

func TestPPrintSlice(t *testing.T) {
	l := []any{1, "sample text", true, 111111, 2222222, 333333, 444444, 555555, 666666, 7777777, 8888888, 99999999}

	var exp = `[1,
 sample text,
 true,
 111111,
 2222222,
 333333,
 444444,
 555555,
 666666,
 7777777,
 8888888,
 99999999]`

	if out := PFormat(l, nil, 1, 80, 2, false, true, false); out != exp {
		t.Errorf("expected 1, got %s", out)
	} else {
		PPrint(l, nil, 1, 80, 2, false, true, false)
	}
}

func TestPPrintMap(t *testing.T) {
	m := map[any]any{
		"1":      11,
		2:        22,
		"sdddd":  22,
		"Sdas":   22,
		"sadsa":  22,
		"sdas":   22,
		"asad":   22,
		"3":      22,
		"2":      22,
		"5555":   22,
		"sadsa1": 22,
		"sdas1":  22,
		"asad1":  22,
		"31":     22,
		"21":     22,
		"55551":  22,
	}
	// 	exp := `{1: 11,
	//  2: 22,
	//  2: 22,
	//  sdas1: 22,
	//  Sdas: 22,
	//  asad: 22,
	//  sadsa1: 22,
	//  21: 22,
	//  sdas: 22,
	//  3: 22,
	//  asad1: 22,
	//  31: 22,
	//  sdddd: 22,
	//  sadsa: 22,
	//  5555: 22,
	//  55551: 22}`
	// 	if out := PFormat(m, nil, 1, 80, 2, false, true, false); out != exp {
	// 		t.Errorf("expected 1, got %s", out)
	// 	} else {
	// 		PPrint(m, nil, 1, 80, 2, false, true, false)
	// 	}
	PPrint(m, nil, 1, 80, 2, false, true, false)
}

func TestPPrintStruct(t *testing.T) {

	sT := createSampleType("sample_text", nil)
	exp := `sampleType(F1=1,
           F2=sample_text2,
           F3=sample_text3,
           F4=sample_text4,
           F5=sample_text5,
           F6=sample_text6,
           F7=sample_text7,
           F8=sample_text8,
           F9=sample_text9,
           F10=sample_text10,
           F11=<nil>,
           private=<private_field>)`
	if out := PFormat(sT, nil, 1, 80, 2, false, true, false); out != exp {
		t.Errorf("expected 1, got %s", out)
	} else {
		PPrint(sT, nil, 1, 80, 2, false, true, false)
	}
}

func TestPPrintNested(t *testing.T) {
	l := []any{1, "sample text", true, 111111, 2222222, 333333, 444444, 555555, 666666, 7777777, 8888888, 99999999}

	nL := []any{
		1, "sample text", true, l,
	}

	nM := map[any]any{
		"1":     11,
		2:       22,
		"slice": nL,
	}

	sT := createSampleType("sample_text", nM)

	exp := `sampleType(F1=1,
           F2=sample_text2,
           F3=sample_text3,
           F4=sample_text4,
           F5=sample_text5,
           F6=sample_text6,
           F7=sample_text7,
           F8=sample_text8,
           F9=sample_text9,
           F10=sample_text10,
           F11={1: 11,
            2: 22,
            slice: [1,
                    sample text,
                    true,
                    [1,
                     sample text,
                     true,
                     111111,
                     2222222,
                     333333,
                     444444,
                     555555,
                     666666,
                     7777777,
                     8888888,
                     99999999]]},
           private=<private_field>)`
	if out := PFormat(sT, nil, 1, 80, 5, false, true, false); out != exp {
		t.Errorf("expected 1, got %s", out)
	} else {
		PPrint(sT, nil, 1, 80, 5, false, true, false)
	}
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
	pp.format(testObject, pp.stream, 2, 0, make(Context), 0)

	// Test with a struct (similar to a dataclass)
	testStruct := struct {
		Name  string
		Value int
	}{
		Name:  "Test Struct",
		Value: 42,
	}
	pp.format(testStruct, pp.stream, 2, 0, make(Context), 0)
}
