# httpio

The `httpio` package provides tools for decoding HTTP requests, decoding url parameters, and encoding HTTP responses in Go, complete with validation rules.

## Getting Started

First, get the package by running:

```sh
go get github.com/cccteam/httpio
```

## Decoder

The Decoder struct is used to decode and validate HTTP requests. It utilizes the json.NewDecoder() function to decode the HTTP request body into a provided struct.

Validation is handled by the Validator interface, which requires a Struct(s interface{}) error function. This function is expected to validate the struct s and return an error if the validation fails.

### Example usage

```go
type MyRequest struct {
    Field1 string `json:"field1" validate:"required"`
    Field2 int    `json:"field2" validate:"required,gt=0"`
}

v := validator.New()

func MyHandler(w http.ResponseWriter, r *http.Request) {
    req := &MyRequest{}
    validatorFunc :=  func(s interface{}) error {

        if err := v.Struct(s); err != nil {
            return err
        }

        return nil
    }

    decoder := httpio.NewDecoder(r, validatorFunc)
    if err := decoder.Decode(req); err != nil {
        // handle error
        return
    }
    // continue processing the request...
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
    // WithParams() middleware should be used to catch parsing errors
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
