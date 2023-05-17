// httpio handles encoding and decoding for http i/o.
// This package is used to standardize request and response handling.
package httpio

//go:generate mockgen -package $GOPACKAGE -destination mock_validator_test.go github.com/cccteam/httpio Validator
//go:generate mockgen -package $GOPACKAGE -destination mock_httpencoder_test.go github.com/cccteam/httpio HTTPEncoder
