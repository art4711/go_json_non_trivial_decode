# More complex decoding/encoding of JSON in Go. #

## Introduction ##

My only complaint about Go is that it's bloody perfect when you follow
the best practice and use the whole package for everything, but things
show cracks when meeting the less than perfectly designed reality of
interacting with the rest of the world that isn't perfect.

So this time we want to decode and encode JSON that has embedded type
information in the objects (ignore the crude emulation of that now, I
just want to see if this is possible without actually digging too deep
into what a real format might look like). We want to provide the
normal interfaces for it, no extra function calls, and we want to
encode/decode normal structs, not stuff adapted to our weird JSON
format.

## The data ##

We start with the top level struct:

```
type xx struct {
	Foo string
	T tface
}
```

Simple enough, only thing worth noting is that Foo has no information
about T. `tface` is in turn defined as:

```
type tface interface {
	Tface() int
}
```

So it can be anything as long as it provides us with the interface we
want. Let's invent two more structs. Their `Tface` functions are there
just to make us provide the interface they are completely irrelevant
otherwise.

```
type t1 struct {
	A int	`json:"a"`
	B int	`json:"b"`
}
func (t t1)Tface() int {
	return t.A + t.B
}

type t2 struct {
	C int		`json:"c"`
	D string	`json:"d"`
}
func (t t2)Tface() int {
	return t.C + len(t.D)
}
```

Now we want a representation of this as JSON:

For t1 we want something like this:

```
{
	"foo": "bar",
	"t": {
		"t1": {
			"a": 1,
			"b": 2
		}
	}
}
```

And for t2:

```
{
	"foo": "bar",
	"t": {
		"t2": {
			"c": 1,
			"d": "str"
		}
	}
}
```

## First implementation ##

The `xx.UnmarshalJSON` function json.Unmarshals the data into an
intermediate struct: `xxJSONdecode` which instead of having the right
type for T has a `map[string]json.RawMessage`. We then brutally look up
which string we got in the map and create the right resulting struct.

The `xx.MarshalJSON` function does the same thing, but with a
different struct.

This is implemented in [revision d3a0a16ada123ebd326e0e8ad92d5c7827774fd6](https://github.com/art4711/go_json_non_trivial_decode/blob/d3a0a16ada123ebd326e0e8ad92d5c7827774fd6/jsm_test.go)

A slightly better approach but on the same theme, is implemented in
[revision e11e04ce654e9be764f2547276558faf9009b158](https://github.com/art4711/go_json_non_trivial_decode/blob/e11e04ce654e9be764f2547276558faf9009b158/jsm_test.go).
We get rid of the separate structs for encoding and decoding and use
slightly more efficient mappings between the type string and the type.

A different approach is implemented in
[revision 707dd70240c85a2f979298820c70dcc130399190](https://github.com/art4711/go_json_non_trivial_decode/blob/707dd70240c85a2f979298820c70dcc130399190/jsm_test.go).
This time the idea is to use slightly less efficient code, but allow
this to scale to arbitrarily many implemented types without regressing
to linear behavior (with a limited number of types this could be quite
slow).  To Unmarshal we use a pre-computed `map[string]reflect.Type`,
then use the type name we get to look up the type, relefect.New() to
create a pointer to a new value of that type, then `.Interface()` to
get an `interface{}` to that pointer, then type assertion back to
`(tface)`. This is the best way I've found of generating values of one
particular interface with dynamically determined types. To Marshal, we
just make out the `tface` interface implement a function that returns
the name of the type. It could be possible to create a map just like
the one we use for Unmarshal, because allegedly reflect.Type is
comparable, but I haven't tried and this works too.


## How to run ##

    go test .

is all that's needed to run this.
