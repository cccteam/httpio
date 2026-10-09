package httpio

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/errors/v5"
)

// jsonTagKey is the struct tag that names a field on the wire.
const jsonTagKey = "json"

// maxBodyPresize bounds how much of a declared Content-Length readBody reserves ahead
// of the read, so a header cannot make the server allocate what the client never sends.
const maxBodyPresize = 1 << 20

// nullLiteral is the JSON null as the key pass holds it.
const nullLiteral = "null"

// ValidatorFunc validates a struct, in whole (Struct) or in the named fields only
// (StructPartial). Each method returns an error when validation fails. A
// *validator.Validate from github.com/go-playground/validator satisfies it.
type ValidatorFunc interface {
	Struct(s any) error
	StructPartial(s any, fields ...string) error
}

// StructDecoder decodes a JSON request body into a plain Request struct and validates
// it. The body is read once and must be a JSON object. Every key it names must reach a
// field of Request: a field with a json tag answers to that tag only; a field without
// one answers to its Go name, exactly or fully lower-cased. A key no field answers to is
// refused with 400 "invalid field in json - <key>", two keys reaching one field with
// 400 "json field name <field> collides with another field name of different case", and
// a JSON null with 400 "<key> cannot be null" unless the field is a pointer, which stays
// nil. A body that is not a JSON object answers 400 "failed to decode request body"; a
// value the field's type cannot hold, 400 "failed to unmarshal request body". With a
// validator (WithValidator), a PATCH validates only the fields the body carried
// (StructPartial, in struct order) and every other method validates the whole struct
// (Struct); a rejection answers 400 "failed validating the request". Every refusal is a
// ClientMessage, written by Encoder.ClientMessage.
//
// A request body bound to a resource decodes with the resource package's decoders,
// which add what the generated struct tags declare (permissions, sizing, nullability,
// former names); a plain request body decodes here.
type StructDecoder[Request any] struct {
	validator ValidatorFunc
	fields    *requestFields
}

// NewStructDecoder creates a StructDecoder for Request, which must be a struct whose
// wire names are distinct.
func NewStructDecoder[Request any]() (*StructDecoder[Request], error) {
	fields, err := newRequestFields(reflect.TypeFor[Request]())
	if err != nil {
		return nil, errors.Wrap(err, "newRequestFields()")
	}

	return &StructDecoder[Request]{
		fields: fields,
	}, nil
}

// WithValidator returns a copy of the decoder that validates every decoded body with v.
func (d *StructDecoder[Request]) WithValidator(v ValidatorFunc) *StructDecoder[Request] {
	decoder := *d
	decoder.validator = v

	return &decoder
}

// Decode decodes the request body into a new Request and validates it.
func (d *StructDecoder[Request]) Decode(request *http.Request) (*Request, error) {
	body, err := readBody(request)
	if err != nil {
		return nil, err
	}

	// The first pass sees every key the body names and whether its value is null, and
	// nothing more: RawMessage keeps it to one scan and a copy of each value, with none
	// of the values built. The second fills the struct from the same bytes.
	keys := make(map[string]json.RawMessage)
	if err := json.Unmarshal(body, &keys); err != nil {
		return nil, NewBadRequestMessageWithError(err, "failed to decode request body")
	}
	if keys == nil {
		// A body of null decodes into a map as nil and into a struct as nothing at all:
		// it is not an object and says nothing about any field.
		return nil, NewBadRequestMessage("failed to decode request body")
	}

	target := new(Request)
	if err := json.Unmarshal(body, target); err != nil {
		return nil, NewBadRequestMessageWithError(err, "failed to unmarshal request body")
	}

	present, err := d.fields.present(keys)
	if err != nil {
		return nil, err
	}

	if err := d.validate(request.Method, target, present); err != nil {
		return nil, err
	}

	return target, nil
}

// readBody reads the whole request body. encoding/json holds a complete value in memory
// before it decodes it, so reading the body once ahead of the two passes costs no memory
// a streaming decode would have saved, and the two passes then share one copy. It also
// ends a body that is a bare scalar (null, true, a number, a string) at EOF, where a
// decoder reading a stream waits for a byte that never comes. A declared Content-Length
// sizes the buffer, up to maxBodyPresize, so a body of known size is read in one
// allocation.
func readBody(request *http.Request) ([]byte, error) {
	var body bytes.Buffer
	if n := request.ContentLength; n > 0 {
		body.Grow(int(min(n, maxBodyPresize)) + bytes.MinRead)
	}
	if _, err := body.ReadFrom(request.Body); err != nil {
		return nil, NewBadRequestMessageWithError(err, "failed to read request body")
	}

	return body.Bytes(), nil
}

// validate runs the validator over the decoded target: the fields the body carried for
// a PATCH, the whole struct for any other method. A nil validator validates nothing.
func (d *StructDecoder[Request]) validate(method string, target *Request, present map[string]struct{}) error {
	if d.validator == nil {
		return nil
	}

	var err error
	if method == http.MethodPatch {
		err = d.validator.StructPartial(target, d.fields.names(present)...)
	} else {
		err = d.validator.Struct(target)
	}
	if err != nil {
		return NewBadRequestMessageWithError(err, "failed validating the request")
	}

	return nil
}

// requestField is one field of the request struct as the decoder reads it.
type requestField struct {
	// name is the Go field name.
	name string
	// acceptsNull is true for a pointer field, the one kind a JSON null may land in.
	acceptsNull bool
}

// requestFields maps the wire names a body may use to the request struct's fields.
type requestFields struct {
	// byName maps each wire name a body may use to the field it reaches: a field's json
	// tag, or without one its Go name and that name lower-cased.
	byName map[string]requestField
	// ordered lists the Go field names in struct order.
	ordered []string
}

// newRequestFields reads the exported fields of t, which must be a struct, refusing a
// wire name that would reach two fields.
func newRequestFields(t reflect.Type) (*requestFields, error) {
	if t.Kind() != reflect.Struct {
		return nil, errors.Newf("request type must be a struct, received %s", t.Kind())
	}

	f := &requestFields{
		byName: make(map[string]requestField),
	}
	for _, field := range reflect.VisibleFields(t) {
		if !field.IsExported() {
			continue
		}
		tag, _, _ := strings.Cut(field.Tag.Get(jsonTagKey), ",")
		if tag == "-" {
			continue
		}

		rf := requestField{
			name:        field.Name,
			acceptsNull: field.Type.Kind() == reflect.Pointer,
		}
		f.ordered = append(f.ordered, field.Name)

		if tag != "" {
			if err := f.add(tag, rf); err != nil {
				return nil, err
			}

			continue
		}

		if err := f.add(field.Name, rf); err != nil {
			return nil, err
		}
		if lower := strings.ToLower(field.Name); lower != field.Name {
			if err := f.add(lower, rf); err != nil {
				return nil, err
			}
		}
	}

	return f, nil
}

// add registers a wire name for a field, refusing one another field already answers to.
func (f *requestFields) add(name string, field requestField) error {
	if _, taken := f.byName[name]; taken {
		return errors.Newf("wire name %s of field %s matches another field", name, field.name)
	}
	f.byName[name] = field

	return nil
}

// lookup finds the field a key reaches, by the key as sent and then lower-cased.
func (f *requestFields) lookup(key string) (requestField, bool) {
	if field, ok := f.byName[key]; ok {
		return field, true
	}
	field, ok := f.byName[strings.ToLower(key)]

	return field, ok
}

// present names the fields a body's keys reach, refusing a key no field answers to, two
// keys reaching one field, and a null into a field that cannot hold one.
func (f *requestFields) present(keys map[string]json.RawMessage) (map[string]struct{}, error) {
	present := make(map[string]struct{}, len(keys))
	for key, value := range keys {
		field, ok := f.lookup(key)
		if !ok {
			return nil, NewBadRequestMessagef("invalid field in json - %s", key)
		}
		if _, seen := present[field.name]; seen {
			return nil, NewBadRequestMessagef("json field name %s collides with another field name of different case", field.name)
		}
		if string(value) == nullLiteral && !field.acceptsNull {
			return nil, NewBadRequestMessagef("%s cannot be null", key)
		}
		present[field.name] = struct{}{}
	}

	return present, nil
}

// names lists the present fields' Go names in struct order.
func (f *requestFields) names(present map[string]struct{}) []string {
	names := make([]string, 0, len(present))
	for _, name := range f.ordered {
		if _, ok := present[name]; ok {
			names = append(names, name)
		}
	}

	return names
}
