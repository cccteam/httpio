package httpio

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/errors/v5"
)

// the Validator interface defines validation methods that are used within the httpio package.
type Validator interface {
	// Struct validates a struct and returns an error if validation fails
	Struct(s interface{}) error
}

// Decoder is a struct that can be used for decoding http requests and validating those requests
type Decoder struct {
	validator Validator
	request   *http.Request
}

// NewDecoder returns a pointer to a new Decoder struct
func NewDecoder(req *http.Request, validator Validator) *Decoder {
	return &Decoder{
		validator: validator,
		request:   req,
	}
}

// Decode parses the http request body and validates it against the struct validation rules
func (d *Decoder) Decode(request interface{}) error {
	if err := json.NewDecoder(d.request.Body).Decode(request); err != nil {
		return errors.Wrap(err, "Decoder.Decode()")
	}

	if err := d.validator.Struct(request); err != nil {
		return errors.Wrap(err, "Validator.Struct()")
	}

	return nil
}
