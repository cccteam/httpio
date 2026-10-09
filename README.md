# httpio

The `httpio` package provides tools for decoding HTTP requests, decoding url parameters, and encoding HTTP responses in Go, complete with validation rules.

## Getting Started

First, get the package by running:

```sh
go get github.com/cccteam/httpio
```

## StructDecoder

`StructDecoder` decodes a JSON request body into a plain struct and validates it. The body must be a JSON object, and every key it names must reach a field of the struct: a field with a `json` tag answers to that tag only; a field without one answers to its Go name, exactly or fully lower-cased. A key no field answers to is refused with a 400 `invalid field in json - <key>`, and a JSON `null` lands only in a pointer field. Every refusal is a `ClientMessage`, so a handler writes it with `Encoder.ClientMessage`.

Validation is optional. `WithValidator` takes a `ValidatorFunc`, an interface with `Struct(s any) error` and `StructPartial(s any, fields ...string) error`, which `*validator.Validate` from `github.com/go-playground/validator` satisfies. A `PATCH` validates only the fields the body carried; every other method validates the whole struct. A rejection is a 400 `failed validating the request`.

A request body bound to a resource decodes with the `resource` package's decoders instead, which add what its generated struct tags declare.

### Example usage

```go
type MyRequest struct {
    Field1 string `json:"field1" validate:"required"`
    Field2 int    `json:"field2" validate:"required,gt=0"`
}

// NewStructDecoder fails only for a type that is not a struct or whose wire names
// collide, a programming error, so construct the decoder once, at startup.
func newDecoder[T any]() *httpio.StructDecoder[T] {
    decoder, err := httpio.NewStructDecoder[T]()
    if err != nil {
        panic(err)
    }

    return decoder.WithValidator(validator.New())
}

func MyHandler() http.HandlerFunc {
    decoder := newDecoder[MyRequest]()

    return func(w http.ResponseWriter, r *http.Request) {
        req, err := decoder.Decode(r)
        if err != nil {
            _ = httpio.NewEncoder(w).ClientMessage(r.Context(), err)

            return
        }
        // continue processing req...
    }
}
```

## Encoder

The `Encoder` struct is used to encode HTTP responses. It has an implementation of the `json.NewEncoder()` function to encode a provided struct into the HTTP response body. The `Encoder` also allows for setting HTTP status codes and headers.

For usage of `Encoder`, please refer to the httpio package's source code.

### Example usage

Here's an example of how to use `Encoder`:

```go
type MyResponse struct {
    Message string `json:"message"`
    Code    int    `json:"code"`
}

func MyHandler(w http.ResponseWriter, r *http.Request) {
    // create response body
    responseBody := &MyResponse{
        Message: "Hello, world!",
        Code:    http.StatusOK,
    }

    // encode and send the response
    if err := httpio.NewEncoder(w).Ok(responseBody); err != nil {
        // handle error
        return
    }
}
```

The `Encoder` struct also provides methods to handle errors and encode HTTP error responses. Here's an example:

```go
func MyHandler(w http.ResponseWriter, r *http.Request) {
    // some operation that may cause an error
    err := someOperation()
    if err != nil {
        // if the operation fails, return an Internal Server Error
        httpio.NewEncoder(w).InternalServerErrorWithMessage("This is what is returned in the response message", err)
        return
    }

    // if the operation is successful, proceed as normal...
}
```

## Params

The Params() generic function serves as an enhancement to the chi router's parameters feature by decoding HTTP URL parameters into native Go types.

Currently the supported types are `string`, `int`, `int64`, `float64`, `bool`, and any type that implements the `encoding.TextUnmarshaler` interface.

### Example usage

```go
// given url: http://myapi.com/api/fileid/26
// and chi route of:          /api/fileid/{fileId}

func MyHandler(w http.ResponseWriter, r *http.Request) {
    param := Param[int64](r, "fileId")
    // param is parsed as type int64
    //
    // WithParams() middleware should be used to catch parsing errors. It responds
    // with a 400 and a generic message naming the parameter; the underlying
    // parse error is logged server-side and never returned to the client.
}
```

## Log

Log returns a `http.HandlerFunc` that logs any error coming from handlers. This provides a more ergonomic feel by allowing errors to be returned from handlers

### Example

```go
func MyHandler() http.HandlerFunc {
	return httpio.Log(func(w http.ResponseWriter, r *http.Request) error {
		// do something
		return errors.New("error")
	})
}
```

## License

This project is licensed under the MIT License.
