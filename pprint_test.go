package pprint

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

type sampleType struct {
	F1      int
	F2      string
	F3      string
	F4      string
	F5      any
	private int
}

// Helper function to generate a SampleType instance
func createSampleType(baseString string, nM any) sampleType {
	return sampleType{
		F1:      1,
		F2:      baseString + "2",
		F3:      baseString + "3",
		F4:      baseString + "4",
		F5:      nM,
		private: 10,
	}
}

func TestPPrintScalars(t *testing.T) {
	i := 1
	ie := "1"
	if out := PFormat(i, nil, 1, 80, 2, false, true, false); out != ie {
		t.Errorf("expected %s, got %s", ie, out)
	} else {
		PPrint(i, nil, 1, 80, 2, false, true, false)
	}
	s := "sample string"
	se := "\"sample string\""
	if out := PFormat(s, nil, 1, 80, 2, false, true, false); out != se {
		t.Errorf("expected %s, got %s", se, out)
	} else {
		PPrint(s, nil, 1, 80, 2, false, true, false)
	}
	f := 0.123
	fe := "0.123"
	if out := PFormat(f, nil, 1, 80, 2, false, true, false); out != fe {
		t.Errorf("expected %s, got %s", fe, out)
	} else {
		PPrint(f, nil, 1, 80, 2, false, true, false)
	}

	b := []byte{15, 20, 30, 40, 50, 100, 15, 20, 30, 40, 50, 100, 15, 20, 30, 40, 50, 100, 15, 20, 30, 40, 50, 100, 15, 20, 30, 40, 50, 100}
	be := `(0f141e2832640f141e283264
0f141e2832640f141e283264
0f141e283264)`
	if out := PFormat(b, nil, 1, 80, 2, false, true, false); out != be {
		t.Errorf("expected %s, got %s", be, out)
	} else {
		PPrint(b, nil, 1, 80, 2, false, true, false)
	}
}

func TestPPrintPtr(t *testing.T) {
	var i *int = new(int)
	*i = 1
	exp := fmt.Sprintf("(%T=%p)&%#v", i, i, *i)
	if out := PFormat(i, nil, 1, 80, 2, false, true, false); out != exp {
		t.Errorf("expected %s, got %s", exp, out)
	} else {
		PPrint(i, nil, 1, 80, 2, false, true, false)
	}
}

func TestPPrintSlice(t *testing.T) {
	l := []any{1, "sample text", true, 111111, 2222222, 333333, 444444, 555555, 666666, 7777777, 8888888, 99999999}

	var exp = `[1,
 "sample text",
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
		t.Errorf("expected %s, got %s", exp, out)
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
	exp := `sampleType(
  F1=1,
  F2="sample_text2",
  F3="sample_text3",
  F4="sample_text4",
  F5=<nil>,
  private=<InaccessibleField>
)`
	if out := PFormat(sT, nil, 1, 80, 2, false, true, false); out != exp {
		t.Errorf("expected %s, got %s", exp, out)
	} else {
		PPrint(sT, nil, 1, 80, 2, false, true, false)
	}
}

func TestPPrintStructPtr(t *testing.T) {
	sT2 := createSampleType("sample_text", nil)
	sT2ptr := &sT2

	sT := createSampleType("sample_text", sT2ptr)
	sTptr := &sT
	exp := fmt.Sprintf(`(*pprint.sampleType=%p)&sampleType(
   F1=1,
   F2="sample_text2",
   F3="sample_text3",
   F4="sample_text4",
   F5=(*pprint.sampleType=%p)&sampleType(
         F1=1,
         F2="sample_text2",
         F3="sample_text3",
         F4="sample_text4",
         F5=<nil>,
         private=<InaccessibleField>
       ),
   private=<InaccessibleField>
 )`, &sT, &sT2)
	if out := PFormat(sTptr, nil, 1, 80, 2, false, true, false); out != exp {
		t.Errorf("expected %s, got %s", exp, out)
	} else {
		PPrint(sTptr, nil, 1, 80, 2, false, true, false)
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

	// exp := `sampleType(F1=1,
	//        F2="sample_text2",
	//        F3="sample_text3",
	//        F4="sample_text4",
	//        F5="sample_text5",
	//        F6="sample_text6",
	//        F7="sample_text7",
	//        F8="sample_text8",
	//        F9="sample_text9",
	//        F10="sample_text10",
	//        F11={"1": 11,
	//         2: 22,
	//         "slice": [1,
	//                   "sample text",
	//                   true,
	//                   [1,
	//                    "sample text",
	//                    true,
	//                    111111,
	//                    2222222,
	//                    333333,
	//                    444444,
	//                    555555,
	//                    666666,
	//                    7777777,
	//                    8888888,
	//                    99999999]]},
	//        private=<InaccessibleField>)`
	// if out := PFormat(sT, nil, 1, 80, 5, false, true, false); out != exp {
	// 	t.Errorf("expected %s, got %s", exp, out)
	// } else {
	PPrint(sT, nil, 1, 80, 5, false, true, false)
	// }
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

func TestRepr(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	p := Person{Name: "John", Age: 30}

	// Using %v
	fmt.Printf("%v\n", p)               // Output: {John 30}
	fmt.Printf("%v\n", "sample_string") // Output: {John 30}

	// Using %#v
	fmt.Println(repr(p))               // Output: main.Person{Name:"John", Age:30}
	fmt.Println(repr("sample_string")) // Output: {John 30}

	l := []any{"121", 1}
	var lf []string
	fmt.Printf("%v\n", l)
	fmt.Printf("%#v\n", l)

	lfmt := "[\n"
	for _, v := range l {
		lf = append(lf, fmt.Sprintf(" %#v", v))
	}
	lfmt += strings.Join(lf, ",\n")
	lfmt += "\n]"
	fmt.Println(lfmt)

}

func TestJsonify(t *testing.T) {
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "John Doe",
			"age":  30,
			"address": struct {
				City    string
				ZipCode string
			}{
				City:    "New York",
				ZipCode: "10001",
			},
			"hobbies": []interface{}{
				"reading",
				"gaming",
				map[string]string{
					"outdoor": "cycling",
				},
			},
		},
		"settings": map[string]interface{}{
			"theme": "dark",
			"notifications": map[string]bool{
				"email": true,
				"sms":   false,
			},
		},
		"stats": struct {
			Posts     int
			Likes     int
			Followers []string
			F         func(any)
		}{
			Posts:     42,
			Likes:     128,
			Followers: []string{"Alice", "Bob", "Charlie"},
			F:         func(a any) { fmt.Println(a) },
		},
	}

	// Convert to JSON
	jsonBytes, err := json.MarshalIndent(data, "", "  ") // Pretty print with indentation
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	// Print JSON string
	fmt.Println(string(jsonBytes))
}

func TempFunc(a any, b int) string {
	return ""
}

func TestMarshalizer(t *testing.T) {
	// Example data with nested structs, maps, and functions
	var ip *float64 = new(float64)
	*ip = 10.0

	pp := sampleType{
		F1: 5,
		F5: ip,
	}
	ppPtr := &pp

	pp2 := sampleType{
		F1: 5, F5: ppPtr,
	}
	ppPtr2 := &pp2

	reg := &SerializersRegistry{}
	var regintf SerializerRegistryInterface = reg

	data := map[any]interface{}{
		"user": map[string]interface{}{
			"name": "John Doe",
			"age":  30,
			"address": struct {
				City    string
				ZipCode string
			}{
				City:    "New York",
				ZipCode: "10001",
			},
			"hobbies": []interface{}{
				"reading",
				"gaming",
				map[string]string{
					"outdoor": "cycling",
				},
			},
		},
		"settings": map[string]interface{}{
			"theme": "dark",
			"notifications": map[string]bool{
				"email": true,
				"sms":   false,
			},
		},
		"stats": struct {
			Posts     int
			Likes     int
			Followers []string
			F         func(any, int)
			P         any
		}{
			Posts:     42,
			Likes:     128,
			Followers: []string{"Alice", "Bob", "Charlie"},
			F:         func(a any, b int) { fmt.Println(a) },
			P:         ppPtr2,
		},
		"stats2": struct {
			F func(any, int) string
		}{
			F: TempFunc,
		},
		"stats3": struct {
			F func(any, int) (string, error)
		}{
			F: func(a any, b int) (string, error) { return "", nil },
		},
		0:              []any{5, 6},
		"marshalizer":  regintf,
		"marshalizer2": regintf.(*SerializersRegistry),
	}

	// Marshal with custom handling
	mr := NewMarshalizer(true, false, false, true)
	jsonBytes, err := mr.Serialize(data)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	// Print JSON string
	fmt.Println(string(jsonBytes))

	PPrint(data, nil, 1, 80, 5, false, true, false)

}

func TestMarshalizerRecursion(t *testing.T) {

	testStr := createSampleType("1", nil)
	ptr := &testStr
	testStr.F5 = ptr

	data := map[any]any{
		"item1": map[any]any{
			"item2": ptr,
			"item3": map[any]any{
				"item4": 5,
			},
		},
	}
	mr := NewMarshalizer(true, false, false, true)
	mr.Serialize(data)

	// context := mr.(*Marshalizer).context
	// fmt.Println(context)

	// var prev MarshalizerContext = make(MarshalizerContext)
	// for key, val := range context {
	// 	if _, exists := prev[key]; exists {
	// 		fmt.Println("Detected recursion", key, val)
	// 	}
	// 	prev[key] = val
	// }

	jsonBytes, err := mr.Serialize(data)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	fmt.Println(string(jsonBytes))

}
