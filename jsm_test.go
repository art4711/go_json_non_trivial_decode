package jsm_test

import (
	"encoding/json"
	"testing"
	"fmt"
)

/*
 * What we want to achieve here is to be able to correctly encode
 * and decode two different json values that look like below.
 * foo is a normal string, the t value contains the type of the
 * value inside, but the struct where we want it represented doesn't
 * want to bother with the type name, it wants the value correctly
 * respresented.
 *
 * We don't want to do any manual json decoding or encoding so we
 * use intermediate structs to do the correct coding and then just
 * copy values. Is there a better way?
 */

// The JSON has no indentation so that we can easily compare
// that what got encoded it the same thing that got decoded
// earlier.
const t1test =`{"foo":"bar","t":{"t1":{"a":1,"b":2}}}`
const t2test = `{"foo":"bar","t":{"t2":{"c":1,"d":"str"}}}`


type tface interface {
	Tface() int
}

type t1 struct {
	A int	`json:"a"`
	B int	`json:"b"`
}
func (t *t1)Tface() int {
	return t.A + t.B
}

type t2 struct {
	C int		`json:"c"`
	D string	`json:"d"`
}
func (t *t2)Tface() int {
	return t.C + len(t.D)
}

type xx struct {
	Foo string `json:"foo"`
	T tface `json:"t"`
}

type xxJSONdecode struct {
	Foo string `json:"foo"`
	T map[string]json.RawMessage `json:"t"`	
}

type xxJSONencode struct {
	Foo string `json:"foo"`
	T map[string]tface `json:"t"`		
}

func (x *xx)UnmarshalJSON(data []byte) error {
	var xd xxJSONdecode

	if err := json.Unmarshal(data, &xd); err != nil {
		return err
	}

	/* Copy the fields that we want copied. */
	x.Foo = xd.Foo

	if len(xd.T) > 1 {
		return fmt.Errorf("More than one t field in JSON data")
	}
	for t, v := range xd.T {
		switch t {
		case "t1":	x.T = &t1{}
		case "t2":	x.T = &t2{}
		default:	return fmt.Errorf("xx.UnmarshalJSON: unknown t type: %v", xd.T)
		}
		return json.Unmarshal(v, x.T)
	}
	return nil
}

func (x *xx)MarshalJSON() ([]byte, error) {
	var xe xxJSONencode
	xe.Foo = x.Foo
	xe.T = make(map[string]tface)
	switch x.T.(type) {
	case *t1:
		xe.T["t1"] = x.T
	case *t2:
		xe.T["t2"] = x.T
	default:
		return nil, fmt.Errorf("xx.MarshalJSON: unknown t type")
	}
	return json.Marshal(&xe)
}

func TestDec1(t *testing.T) {
	var totest xx
	if err := json.Unmarshal([]byte(t1test), &totest); err != nil {
		t.Fatal(err)
	}
	if totest.Foo != "bar" {
		t.Errorf("foo mismatch: %s != bar", totest.Foo)
	}
	xt1, ok := totest.T.(*t1)
	if !ok {
		t.Errorf("foo.t wrong type: %V", totest.T)
	}
	if xt1.A != 1 || xt1.B != 2 {
		t.Errorf("foo.t bad value: %v", xt1)
	}
}

func TestDec2(t *testing.T) {
	var totest xx
	if err := json.Unmarshal([]byte(t2test), &totest); err != nil {
		t.Fatal(err)
	}
	if totest.Foo != "bar" {
		t.Errorf("foo mismatch: %s != bar", totest.Foo)
	}
	xt2, ok := totest.T.(*t2)
	if !ok {
		t.Errorf("foo.t wrong type: %V", totest.T)
	}
	if xt2.C != 1 || xt2.D != "str" {
		t.Errorf("foo.t bad value: %v", xt2)
	}
}

func TestEnc1(t *testing.T) {
	var totest xx
	if err := json.Unmarshal([]byte(t1test), &totest); err != nil {
		t.Fatal(err)
	}
	bs, err := json.Marshal(&totest)
	if err != nil {
		t.Fatal(err)
	}
	s := string(bs)
	if s != t1test {
		t.Errorf("encoding '%s' != original '%s'", s, t1test);
	}
}

func TestEnc2(t *testing.T) {
	var totest xx
	if err := json.Unmarshal([]byte(t2test), &totest); err != nil {
		t.Fatal(err)
	}
	bs, err := json.Marshal(&totest)
	if err != nil {
		t.Fatal(err)
	}
	s := string(bs)
	if s != t2test {
		t.Errorf("encoding '%s' != original '%s'", s, t2test);
	}
}