package rack_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stevecallear/rack"
)

func TestStatusCode(t *testing.T) {
	err := errors.New("error")

	tests := []struct {
		name string
		err  error
		exp  int
	}{
		{
			name: "should return 500 if the error is not a status error",
			err:  err,
			exp:  http.StatusInternalServerError,
		},
		{
			name: "should return status error codes",
			err:  rack.WrapError(http.StatusBadRequest, err),
			exp:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			act := rack.StatusCode(tt.err)
			if act != tt.exp {
				t.Errorf("got %d, expected %d", act, tt.exp)
			}
		})
	}
}

func TestStatusError_Code(t *testing.T) {
	t.Run("should return the status code", func(t *testing.T) {
		const exp = http.StatusConflict
		err := rack.WrapError(exp, errors.New("error"))

		act := err.Code()
		if act != exp {
			t.Errorf("got %d, expected %d", act, exp)
		}
	})
}

func TestStatusError_Error(t *testing.T) {
	t.Run("should return the wrapped error message", func(t *testing.T) {
		const exp = "error"
		sut := rack.WrapError(http.StatusBadRequest, errors.New(exp))

		act := sut.Error()
		if act != exp {
			t.Errorf("got %s, expected %s", act, exp)
		}
	})
}

func TestStatusError_Unwrap(t *testing.T) {
	t.Run("should return the wrapped error", func(t *testing.T) {
		exp := errors.New("error")
		sut := rack.WrapError(http.StatusBadRequest, exp)

		act := errors.Unwrap(sut)
		if act != exp {
			t.Errorf("got %v, expected %v", act, exp)
		}
	})
}
