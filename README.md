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
type for T has a map[string]json.RawMessage. We then brutally look up
which string we got in the map and create the right resulting struct.

The `xx.MarshalJSON` function does the same thing, but with a
different struct.

This is implemented in revision d3a0a16ada123ebd326e0e8ad92d5c7827774fd6
