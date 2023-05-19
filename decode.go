package httpio

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/errors/v5"
)

// ValidatorFunc is a function that validates s
// It returns an error if the validation fails
type ValidatorFunc func(s interface{}) error

// Decoder is a struct that can be used for decoding http requests and validating those requests
type Decoder struct {
	validateFunc ValidatorFunc
	request      *http.Request
}

// NewDecoder returns a pointer to a new Decoder struct
func NewDecoder(req *http.Request, validator ValidatorFunc) *Decoder {
	return &Decoder{
		validateFunc: validator,
		request:      req,
	}
}

// Decode parses the http request body and validates it against the struct validation rules
func (d *Decoder) Decode(request interface{}) error {
	if err := json.NewDecoder(d.request.Body).Decode(request); err != nil {
		return errors.Wrap(err, "Decoder.Decode()")
	}

	if err := d.validateFunc(request); err != nil {
		return errors.Wrap(err, "failed validating the request")
	}

	return nil
}
