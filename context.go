package rack

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
)

type (
	// Context represents a handler context
	Context interface {
		// Context returns the function invocation context
		Context() context.Context

		// Request returns the canonical request
		Request() *Request

		// Response returns the canonical response
		Response() *Response

		// Get returns the stored value with the specified key
		Get(key string) interface{}

		// Set stores the specified value in the context
		Set(key string, v interface{})

		// Path returns the path parameter with the specified key
		// An empty string is returned if no parameter exists.
		Path(key string) string

		// Query returns the first query string parameter with the specified key
		// An empty string is returned if no parameter exists. If all query string values
		// are required, then the raw values can be accessed using Request().Query[key].
		Query(key string) string

		// Bind unmarshals the request body into the specified value
		// Currently only JSON request bodies are supported.
		Bind(v interface{}) error

		// NoContent writes the specified status code to the response without a body
		NoContent(code int) error

		// String writes the specified status code and value to the response
		String(code int, s string) error

		// JSON writes the specified status code and value to the response as JSON
		JSON(code int, v interface{}) error
	}

	handlerContext struct {
		ctx      context.Context
		store    map[string]interface{}
		request  *Request
		response *Response
		onBind   func(Context, interface{}) error
		mu       *sync.RWMutex
	}
)

func (c *handlerContext) Context() context.Context {
	return c.ctx
}

func (c *handlerContext) Request() *Request {
	return c.request
}

func (c *handlerContext) Response() *Response {
	return c.response
}

func (c *handlerContext) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.store != nil {
		return c.store[key]
	}

	return nil
}

func (c *handlerContext) Set(key string, v interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.store == nil {
		c.store = map[string]interface{}{key: v}
	} else {
		c.store[key] = v
	}
}

func (c *handlerContext) Path(key string) string {
	return c.request.Path[key]
}

func (c *handlerContext) Query(key string) string {
	return c.request.Query.Get(key)
}

func (c *handlerContext) Bind(v interface{}) error {
	if c.request.Body == "" {
		return nil
	}

	err := json.Unmarshal([]byte(c.request.Body), v)
	if err != nil {
		return WrapError(http.StatusBadRequest, err)
	}

	return c.onBind(c, v)
}

func (c *handlerContext) NoContent(code int) error {
	c.response.StatusCode = code
	return nil
}

func (c *handlerContext) String(code int, s string) error {
	c.response.StatusCode = code
	c.response.Body = s
	c.response.Headers["Content-Type"] = []string{"text/plain"}

	return nil
}

func (c *handlerContext) JSON(code int, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	c.response.StatusCode = code
	c.response.Body = string(b)
	c.response.Headers["Content-Type"] = []string{"application/json"}

	return nil
}
