package jsm_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
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
const t1test = `{"foo":"bar","t":{"t1":{"a":1,"b":2}}}`
const t2test = `{"foo":"bar","t":{"t2":{"c":1,"d":"str"}}}`

type tface interface {
	TypName() string
}

type t1 struct {
	A int `json:"a"`
	B int `json:"b"`
}

func (t *t1) TypName() string {
	return "t1"
}

type t2 struct {
	C int    `json:"c"`
	D string `json:"d"`
}

func (t *t2) TypName() string {
	return "t2"
}

type xx struct {
	Foo string `json:"foo"`
	T   tface  `json:"t"`
}

type xxJSON struct {
	Foo string                      `json:"foo"`
	T   map[string]*json.RawMessage `json:"t"`
}

var xxtypes = map[string]reflect.Type{
	"t1": reflect.TypeOf(t1{}),
	"t2": reflect.TypeOf(t2{}),
}

func (x *xx) UnmarshalJSON(data []byte) error {
	var xd xxJSON

	if err := json.Unmarshal(data, &xd); err != nil {
		return err
	}

	/* Copy the fields that we want copied. */
	x.Foo = xd.Foo

	if len(xd.T) > 1 {
		return fmt.Errorf("More than one t field in JSON data")
	}
	for t, v := range xd.T {
		typ, ok := xxtypes[t]
		if !ok {
			return fmt.Errorf("xx.UnmarshalJSON: unknown t type: %v", xd.T)
		}
		x.T = reflect.New(typ).Interface().(tface)
		return json.Unmarshal(*v, x.T)
	}
	return nil
}

func (x *xx) MarshalJSON() ([]byte, error) {
	var xd xxJSON

	xd.Foo = x.Foo
	j, err := json.Marshal(x.T)
	if err != nil {
		return nil, err
	}
	xd.T = make(map[string]*json.RawMessage)
	jr := json.RawMessage(j)
	xd.T[x.T.TypName()] = &jr

	return json.Marshal(&xd)
}

// This is testing how json.RawMessage works. I needed to figure out that
// json.RawMessage only has a pointer receivers, so even when it's part of a
// struct it will have to a pointer value, otherwise it only gets encoded
// as a byte array and as such it gets base64 encoded insted of passed through
// as we want. This was slightly confusing.
func TestRawMessage(t *testing.T) {
	msg := `{"foo":"bar"}`
	var m json.RawMessage
	if err := json.Unmarshal([]byte(msg), &m); err != nil {
		t.Fatal(err)
	}
	msg2, err := json.Marshal(&m) // if this isn't &m, we're screwed.
	if err != nil {
		t.Fatal(err)
	}
	if msg != string(msg2) {
		t.Errorf("wtf: %s != %s", msg, msg2)
	}
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
		t.Errorf("encoding '%s' != original '%s'", s, t1test)
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
		t.Errorf("encoding '%s' != original '%s'", s, t2test)
	}
}
