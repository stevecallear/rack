package rack_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stevecallear/rack"
)

func TestStatusCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		exp  int
	}{
		{
			name: "should return 500 if the error is not a status error",
			err:  errors.New("error"),
			exp:  http.StatusInternalServerError,
		},
		{
			name: "should return status error codes",
			err:  rack.NewError(http.StatusBadRequest, "error"),
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
		err := rack.NewError(exp, "error")

		act := err.Code()
		if act != exp {
			t.Errorf("got %d, expected %d", act, exp)
		}
	})
}

func TestStatusError_Error(t *testing.T) {
	tests := []struct {
		name string
		sut  *rack.StatusError
		exp  string
	}{
		{
			name: "should return the error message",
			sut:  rack.NewError(http.StatusBadRequest, "error"),
			exp:  "error",
		}, {
			name: "should return the wrapped error message",
			sut:  rack.WrapError(http.StatusBadRequest, errors.New("error")),
			exp:  "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			act := tt.sut.Error()
			if act != tt.exp {
				t.Errorf("got %s, expected %s", act, tt.exp)
			}
		})
	}
}

func TestStatusError_Unwrap(t *testing.T) {
	err := errors.New("error")

	tests := []struct {
		name string
		sut  *rack.StatusError
		exp  error
	}{
		{
			name: "should return nil if there is no wrapped error",
			sut:  rack.NewError(http.StatusBadRequest, "error"),
			exp:  nil,
		}, {
			name: "should return the wrapped error",
			sut:  rack.WrapError(http.StatusBadRequest, err),
			exp:  err,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			act := tt.sut.Unwrap()
			if act != tt.exp {
				t.Errorf("got %v, expected %v", act, tt.exp)
			}
		})
	}
}
