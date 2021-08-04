package rack

import (
	"context"
	"net/http"
	"net/url"
	"sync"

	"github.com/aws/aws-lambda-go/lambda"
)

type (
	// HandlerFunc represents a handler function
	HandlerFunc func(Context) error

	// MiddlewareFunc represents a middleware function
	MiddlewareFunc func(HandlerFunc) HandlerFunc

	// Config represent handler configuration
	Config struct {
		Resolver        Resolver
		Middleware      MiddlewareFunc
		OnBind          func(Context, interface{}) error
		OnError         func(Context, error) error
		OnEmptyResponse HandlerFunc
	}

	// Request represents a canonical request type
	Request struct {
		Method  string
		RawPath string
		Path    map[string]string
		Query   url.Values
		Header  http.Header
		Body    string
		Event   interface{}
	}

	// Response represents a canonical response type
	Response struct {
		StatusCode int
		Headers    http.Header
		Body       string
	}

	invokeFunc func(context.Context, []byte) ([]byte, error)
)

// New returns a new lambda handler for the specified function
func New(h HandlerFunc) lambda.Handler {
	return NewWithConfig(Config{}, h)
}

// NewWithConfig returns a new lambda handler for the specified function and configuration
func NewWithConfig(c Config, h HandlerFunc) lambda.Handler {
	if c.Middleware != nil {
		h = c.Middleware(h)
	}

	resolver := c.Resolver
	if resolver == nil {
		resolver = defaultResolver
	}

	onError := c.OnError
	if onError == nil {
		onError = defaultErrorHandler
	}

	onBind := c.OnBind
	if onBind == nil {
		onBind = func(Context, interface{}) error { return nil }
	}

	onEmptyResponse := c.OnEmptyResponse
	if onEmptyResponse == nil {
		onEmptyResponse = func(c Context) error {
			return c.NoContent(http.StatusOK)
		}
	}

	return invokeFunc(func(ctx context.Context, payload []byte) ([]byte, error) {
		p, err := resolver.Resolve(payload)
		if err != nil {
			return nil, err
		}

		req, err := p.UnmarshalRequest(payload)
		if err != nil {
			return nil, err
		}

		c := &handlerContext{
			ctx:     ctx,
			request: req,
			response: &Response{
				Headers: http.Header{},
			},
			onBind: onBind,
			mu:     new(sync.RWMutex),
		}

		if err = h(c); err != nil {
			if err = onError(c, err); err != nil {
				return nil, err
			}
		}

		if c.response.StatusCode == 0 {
			if err = onEmptyResponse(c); err != nil {
				if err = onError(c, err); err != nil {
					return nil, err
				}
			}
		}

		return p.MarshalResponse(c.response)
	})
}

// Chain returns a middleware func that chains the specified funcs
func Chain(m ...MiddlewareFunc) MiddlewareFunc {
	return MiddlewareFunc(func(n HandlerFunc) HandlerFunc {
		for i := len(m) - 1; i >= 0; i-- {
			n = m[i](n)
		}
		return n
	})
}

func (fn invokeFunc) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	return fn(ctx, payload)
}

func defaultErrorHandler(c Context, err error) error {
	res := struct {
		Message string `json:"message"`
	}{
		Message: err.Error(),
	}

	return c.JSON(StatusCode(err), &res)
}
