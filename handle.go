package httpio

import (
	"net/http"
	"strings"

	"github.com/cccteam/logger"
	"github.com/go-playground/errors/v5"
)

// Handle returns a http.HandlerFunc that logs any error coming from handlers.
// This provides a more ergonomic feel by allowing errors to be returned from handlers
// 
// Example usage:
//
// 	func Handler() http.HandlerFunc {
// 		return httpio.Handle(func(w http.ResponseWriter, r *http.Request) error {
// 			// do something
// 			return errors.New("error")
// 		})
// 	}
func Handle(handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			cerr := &ClientMessage{}
			if errors.As(err, &cerr) {
				messages := strings.Join(Messages(err), "', '")
				if cerr.msgType < http.StatusInternalServerError {
					logger.Req(r).Info(cerr)
					if messages != "" {
						logger.Req(r).Infof("['%s']", messages)
					}
				} else {
					logger.Req(r).Error(cerr)
					if messages != "" {
						logger.Req(r).Errorf("['%s']", messages)
					}
				}
			}
		}
	})
}
