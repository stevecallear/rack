package rack_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"

	"github.com/stevecallear/rack"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		handler rack.HandlerFunc
		payload []byte
		exp     []byte
		err     bool
	}{
		{
			name:    "should return an error if the payload is invalid",
			payload: []byte("{"),
			err:     true,
		},
		{
			name: "should handle handler errors",
			handler: func(rack.Context) error {
				return rack.WrapError(http.StatusConflict, errors.New("error"))
			},
			payload: newV2Request(nil),
			exp: newV2Response(func(r *events.APIGatewayV2HTTPResponse) {
				r.StatusCode = http.StatusConflict
				r.Headers = map[string]string{
					"Content-Type": "application/json",
				}
				r.MultiValueHeaders = map[string][]string{
					"Content-Type": {"application/json"},
				}
				r.Body = `{"message":"error"}`
			}),
		},
		{
			name: "should return a default response if none is written",
			handler: func(rack.Context) error {
				return nil
			},
			payload: newV2Request(nil),
			exp: newV2Response(func(r *events.APIGatewayV2HTTPResponse) {
				r.StatusCode = http.StatusOK
			}),
		},
		{
			name: "should return the response",
			handler: func(c rack.Context) error {
				v := struct {
					Key string `json:"key"`
				}{}

				if err := c.Bind(&v); err != nil {
					return err
				}

				return c.String(http.StatusOK, v.Key)
			},
			payload: newV2Request(func(r *events.APIGatewayV2HTTPRequest) {
				r.Body = `{"key":"value"}`
			}),
			exp: newV2Response(func(r *events.APIGatewayV2HTTPResponse) {
				r.Headers = map[string]string{
					"Content-Type": "text/plain",
				}
				r.MultiValueHeaders = map[string][]string{
					"Content-Type": {"text/plain"},
				}
				r.Body = "value"
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := rack.New(tt.handler)

			act, err := h.Invoke(context.Background(), tt.payload)

			assertErrorExists(t, err, tt.err)
			assertDeepEqual(t, act, tt.exp)
		})
	}
}

func TestNewWithConfig(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*rack.Config)
		handler rack.HandlerFunc
		payload []byte
		exp     []byte
		err     bool
	}{
		{
			name: "should return an error if the payload is invalid",
			setup: func(c *rack.Config) {
				c.Resolver = rack.ResolveStatic(rack.APIGatewayV2HTTPEventProcessor)
			},
			payload: []byte("{"),
			err:     true,
		},
		{
			name: "should use the error handler",
			setup: func(c *rack.Config) {
				c.OnError = func(_ rack.Context, err error) error {
					return err
				}
			},
			handler: func(c rack.Context) error {
				return errors.New("error")
			},
			payload: newV2Request(nil),
			err:     true,
		},
		{
			name: "should use the bind callback",
			setup: func(c *rack.Config) {
				c.OnBind = func(rack.Context, interface{}) error {
					return errors.New("error")
				}
			},
			handler: func(c rack.Context) error {
				v := struct {
					Key string `json:"key"`
				}{}
				return c.Bind(&v)
			},
			payload: newV2Request(func(r *events.APIGatewayV2HTTPRequest) {
				r.Body = `{"key":"value"}`
			}),
			exp: newV2Response(func(r *events.APIGatewayV2HTTPResponse) {
				r.StatusCode = http.StatusInternalServerError
				r.Headers = map[string]string{
					"Content-Type": "application/json",
				}
				r.MultiValueHeaders = map[string][]string{
					"Content-Type": {"application/json"},
				}
				r.Body = `{"message":"error"}`
			}),
		},
		{
			name: "should use the empty response handler",
			setup: func(c *rack.Config) {
				c.OnError = func(_ rack.Context, err error) error {
					return err
				}
				c.OnEmptyResponse = func(rack.Context) error {
					return errors.New("error")
				}
			},
			handler: func(c rack.Context) error {
				return nil
			},
			payload: newV2Request(nil),
			err:     true,
		},
		{
			name: "should use the middleware",
			setup: func(c *rack.Config) {
				c.Middleware = func(n rack.HandlerFunc) rack.HandlerFunc {
					return func(c rack.Context) error {
						c.Response().Headers.Set("X-Custom-Header", "header")
						return n(c)
					}
				}
			},
			handler: func(c rack.Context) error {
				return c.String(http.StatusOK, "body")
			},
			payload: newV2Request(nil),
			exp: newV2Response(func(r *events.APIGatewayV2HTTPResponse) {
				r.Headers = map[string]string{
					"Content-Type":    "text/plain",
					"X-Custom-Header": "header",
				}
				r.MultiValueHeaders = map[string][]string{
					"Content-Type":    {"text/plain"},
					"X-Custom-Header": {"header"},
				}
				r.Body = "body"
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c rack.Config
			tt.setup(&c)

			h := rack.NewWithConfig(c, tt.handler)

			act, err := h.Invoke(context.Background(), tt.payload)

			assertErrorExists(t, err, tt.err)
			assertDeepEqual(t, act, tt.exp)
		})
	}
}

func TestChain(t *testing.T) {
	mw := func(sb *strings.Builder, s string) rack.MiddlewareFunc {
		return func(n rack.HandlerFunc) rack.HandlerFunc {
			return func(c rack.Context) error {
				sb.WriteString(s)
				defer sb.WriteString(s)
				return n(c)
			}
		}
	}

	t.Run("should chain the functions", func(t *testing.T) {
		sb := new(strings.Builder)
		sut := rack.Chain(mw(sb, "1"), mw(sb, "2"))

		h := sut(func(c rack.Context) error {
			sb.WriteString("h")
			return nil
		})

		err := h(nil)
		assertErrorExists(t, err, false)

		if act, exp := sb.String(), "12h21"; act != exp {
			t.Errorf("got %s, expected %s", act, exp)
		}
	})
}

func marshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func unmarshal(b []byte, v interface{}) interface{} {
	if err := json.Unmarshal(b, v); err != nil {
		panic(err)
	}
	return v
}

func newV2Request(fn func(*events.APIGatewayV2HTTPRequest)) []byte {
	r := &events.APIGatewayV2HTTPRequest{
		Version: "2.0",
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			APIID: "apiid",
		},
	}

	if fn != nil {
		fn(r)
	}

	b, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}

	return b
}

func newV2Response(fn func(*events.APIGatewayV2HTTPResponse)) []byte {
	r := &events.APIGatewayV2HTTPResponse{
		StatusCode:        http.StatusOK,
		Headers:           map[string]string{},
		MultiValueHeaders: map[string][]string{},
		Cookies:           []string{},
	}

	if fn != nil {
		fn(r)
	}

	b, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}

	return b
}

func assertErrorExists(t *testing.T, act error, exp bool) {
	if act != nil && !exp {
		t.Errorf("got %v, expected nil", act)
	}
	if act == nil && exp {
		t.Error("got nil, expected an error")
	}
}

func assertDeepEqual(t *testing.T, act, exp interface{}) {
	if !reflect.DeepEqual(act, exp) {
		t.Errorf("got %v, expected %v", act, exp)
	}
}
