package apix

import (
	"fmt"
	"net/http"
)

func WithRecover(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				WriteError(w, http.StatusBadRequest, recoverMessage(recovered))
			}
		}()

		next(w, r)
	}
}

func recoverMessage(recovered any) string {
	switch value := recovered.(type) {
	case error:
		return value.Error()
	case string:
		return value
	default:
		return fmt.Sprint(value)
	}
}
