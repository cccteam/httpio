package httpio

import (
	"context"
	"net/http"
	"strings"

	"github.com/cccteam/logger"
	"github.com/go-playground/errors/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Log returns a http.HandlerFunc that logs any error coming from handlers.
// This provides a more ergonomic feel by allowing errors to be returned from handlers
//
// Example usage:
//
//	func Handler() http.HandlerFunc {
//		return httpio.Log(func(w http.ResponseWriter, r *http.Request) error {
//			// do something
//			return errors.New("error")
//		})
//	}
func Log(handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err == nil {
			return
		}

		cerr := &ClientMessage{}
		isClientErr := errors.As(err, &cerr)

		if r.Context().Err() != nil && (errors.Is(err, context.Canceled) || status.Code(err) == codes.Canceled || strings.Contains(err.Error(), "context canceled")) {
			if !isClientErr || cerr.msgType != clientClosedRequest {
				err = NewClientClosedRequestWithError(err)
				isClientErr = errors.As(err, &cerr)
			}
		}

		if !isClientErr {
			logger.FromReq(r).Error(err)

			return
		}

		messages := strings.Join(Messages(err), "', '")
		if cerr.msgType < internalServerError {
			logger.FromReq(r).Info(err)
			if messages != "" {
				logger.FromReq(r).Infof("messages=['%s']", messages)
			}
		} else {
			logger.FromReq(r).Error(err)
			if messages != "" {
				logger.FromReq(r).Errorf("messages=['%s']", messages)
			}
		}
	}
}
