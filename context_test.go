package rack_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"

	"github.com/stevecallear/rack"
)

func TestContext_Context(t *testing.T) {
	t.Run("should return the context", func(t *testing.T) {
		exp := context.Background()

		h := rack.New(func(c rack.Context) error {
			act := c.Context()
			if act != exp {
				t.Errorf("got %v, expected %v", act, exp)
			}
			return nil
		})

		h.Invoke(exp, newV2Request(nil))
	})
}

func TestContext_Request(t *testing.T) {
	t.Run("should return the request", func(t *testing.T) {
		const exp = "expected"

		p := newV2Request(func(r *events.APIGatewayV2HTTPRequest) {
			r.Body = exp
		})

		h := rack.New(func(c rack.Context) error {
			if c.Request().Body != exp {
				t.Errorf("got %s, expected %s", c.Request().Body, exp)
			}
			return nil
		})

		h.Invoke(context.Background(), p)
	})
}

func TestContext_Get(t *testing.T) {
	tests := []struct {
		name  string
		setup func(rack.Context)
		exp   interface{}
	}{
		{
			name:  "should return nil if the key does not exist",
			setup: func(rack.Context) {},
			exp:   nil,
		},
		{
			name: "should return the value",
			setup: func(c rack.Context) {
				c.Set("key", "value")
				c.Set("other", "other")
			},
			exp: "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := rack.New(func(c rack.Context) error {
				tt.setup(c)
				act := c.Get("key")

				if act != tt.exp {
					t.Errorf("got %v, expected %v", act, tt.exp)
				}

				return nil
			})

			_, err := h.Invoke(context.Background(), newV2Request(nil))
			assertErrorExists(t, err, false)
		})
	}
}

func TestContext_Path(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		exp     string
	}{
		{
			name:    "should return empty if the path parameters are nil",
			payload: newV2Request(nil),
			exp:     "",
		},
		{
			name: "should return empty if the path key does not exist",
			payload: newV2Request(func(r *events.APIGatewayV2HTTPRequest) {
				r.PathParameters = map[string]string{"other": "value"}
			}),
			exp: "",
		},
		{
			name: "should return the path parameter",
			payload: newV2Request(func(r *events.APIGatewayV2HTTPRequest) {
				r.PathParameters = map[string]string{"key": "value"}
			}),
			exp: "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := rack.New(func(c rack.Context) error {
				act := c.Path("key")
				if act != tt.exp {
					t.Errorf("got %s, expected %s", act, tt.exp)
				}

				return nil
			})

			_, err := h.Invoke(context.Background(), tt.payload)
			if err != nil {
				t.Errorf("got %v, expected nil", err)
			}
		})
	}
}

func TestContext_Query(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		exp     string
	}{
		{
			name:    "should return empty if the query string parameters are nil",
			payload: newV2Request(nil),
			exp:     "",
		},
		{
			name: "should return empty if the query string does not exist",
			payload: newV2Request(func(r *events.APIGatewayV2HTTPRequest) {
				r.QueryStringParameters = map[string]string{"other": "value"}
			}),
			exp: "",
		},
		{
			name: "should return the query string value",
			payload: newV2Request(func(r *events.APIGatewayV2HTTPRequest) {
				r.QueryStringParameters = map[string]string{"key": "value"}
			}),
			exp: "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := rack.New(func(c rack.Context) error {
				act := c.Query("key")
				if act != tt.exp {
					t.Errorf("got %s, expected %s", act, tt.exp)
				}

				return nil
			})

			_, err := h.Invoke(context.Background(), tt.payload)
			if err != nil {
				t.Errorf("got %v, expected nil", err)
			}
		})
	}
}

func TestContext_Bind(t *testing.T) {
	type obj struct {
		Key string `json:"key"`
	}

	body := obj{Key: "value"}

	tests := []struct {
		name    string
		payload []byte
		exp     obj
		err     bool
	}{
		{
			name:    "should do nothing if the body is empty",
			payload: newV2Request(nil),
			exp:     obj{},
		},
		{
			name: "should return an error if the body is invalid",
			payload: newV2Request(func(r *events.APIGatewayV2HTTPRequest) {
				r.Body = "{"
			}),

			err: true,
		},
		{
			name: "should bind the body",
			payload: newV2Request(func(r *events.APIGatewayV2HTTPRequest) {
				r.Body = string(marshal(&body))
			}),
			exp: body,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := rack.New(func(c rack.Context) error {
				var act obj
				err := c.Bind(&act)

				assertErrorExists(t, err, tt.err)
				assertDeepEqual(t, act, tt.exp)

				return nil
			})

			_, err := h.Invoke(context.Background(), tt.payload)
			assertErrorExists(t, err, false)
		})
	}
}

func TestContext_NoContent(t *testing.T) {
	t.Run("should set the status code", func(t *testing.T) {
		exp := &events.APIGatewayV2HTTPResponse{
			StatusCode:        http.StatusCreated,
			Headers:           map[string]string{},
			MultiValueHeaders: map[string][]string{},
			Cookies:           []string{},
		}

		h := rack.New(func(c rack.Context) error {
			return c.NoContent(exp.StatusCode)
		})

		b, err := h.Invoke(context.Background(), newV2Request(nil))
		assertErrorExists(t, err, false)

		act := new(events.APIGatewayV2HTTPResponse)
		unmarshal(b, act)

		assertDeepEqual(t, *act, *exp)
	})
}

func TestContext_String(t *testing.T) {
	t.Run("should set the status code and body", func(t *testing.T) {
		exp := &events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusOK,
			Body:       "value",
			Headers: map[string]string{
				"Content-Type": "text/plain",
			},
			MultiValueHeaders: map[string][]string{
				"Content-Type": {"text/plain"},
			},
			Cookies: []string{},
		}

		h := rack.New(func(c rack.Context) error {
			return c.String(exp.StatusCode, exp.Body)
		})

		b, err := h.Invoke(context.Background(), newV2Request(nil))
		assertErrorExists(t, err, false)

		act := new(events.APIGatewayV2HTTPResponse)
		unmarshal(b, act)

		assertDeepEqual(t, *act, *exp)
	})
}

func TestContext_JSON(t *testing.T) {
	type obj struct {
		Key string `json:"key"`
	}

	tests := []struct {
		name    string
		handler rack.HandlerFunc
		exp     []byte
		err     bool
	}{
		{
			name: "should return marshal errors",
			handler: func(c rack.Context) error {
				return c.JSON(http.StatusOK, make(chan struct{}))
			},
			err: true,
		},
		{
			name: "should set the status code and body",
			handler: func(c rack.Context) error {
				return c.JSON(http.StatusCreated, &obj{Key: "value"})
			},
			exp: newV2Response(func(r *events.APIGatewayV2HTTPResponse) {
				r.StatusCode = http.StatusCreated
				r.Body = `{"key":"value"}`
				r.Headers = map[string]string{
					"Content-Type": "application/json",
				}
				r.MultiValueHeaders = map[string][]string{
					"Content-Type": {"application/json"},
				}
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := rack.NewWithConfig(rack.Config{
				OnError: func(_ rack.Context, err error) error {
					return err
				},
			}, tt.handler)

			act, err := h.Invoke(context.Background(), newV2Request(nil))
			assertErrorExists(t, err, tt.err)
			assertDeepEqual(t, act, tt.exp)
		})
	}
}
