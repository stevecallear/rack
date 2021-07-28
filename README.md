# Rack
[![Build Status](https://github.com/stevecallear/rack/actions/workflows/build.yml/badge.svg)](https://github.com/stevecallear/rack/actions/workflows/build.yml)
[![codecov](https://codecov.io/gh/stevecallear/rack/branch/master/graph/badge.svg)](https://codecov.io/gh/stevecallear/rack)
[![Go Report Card](https://goreportcard.com/badge/github.com/stevecallear/rack)](https://goreportcard.com/report/github.com/stevecallear/rack)

Rack provides an opinionated wrapper for AWS Lambda handlers written in Go. The concept is similar to that offered by [chop](https://github.com/stevecallear/chop) and [aws-lambda-go-api-proxy](https://github.com/awslabs/aws-lambda-go-api-proxy), but without the integration with the standard HTTP modules.

The intention of the module is to remove a lot of the boilerplate involved in writing handler functions for scenarios that do not make use of HTTP routing. Typically this would be when an individual Lambda function is deployed for each resource in an API as opposed to using a router within a single function.

## Getting Started
```
go get github.com/stevecallear/rack
```

```
import (
    "github.com/aws/aws-lambda-go/lambda"
	"github.com/stevecallear/rack"
)

func main() {
    h := rack.New(func(c rack.Context) error {
		t, err := store.GetTask(c.Context(), c.Path("id"))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, &t)
	})

    lambda.StartHandler(h)
}
```

## Handler
A handler must satisfy the `func(rack.Context) error` signature. The supplied `Context` provides a number of request accessors and response writers for common operations. 

Operations not available on the `Context` can be performed by accessing the canonical request and response objects using `Request` and `Response` respectively.
```
h := rack.New(func(c rack.Context) error {
    v := c.Request().Header.Get("X-Custom-Header")
    c.Response().Header.Set("X-Custom-Header", v)

    return c.NoContent(http.StatusOK)
})
```

The incoming event and Lamdba context are also available if required. The following example assumes that the event type is guaranteed. A type switch or equivalent should be used if the handler is handling multiple event types.
```
h := rack.NewWithConfig(cfg, func(c rack.Context) error {
    e := c.Request().Event.(*events.APIGatewayV2HTTPRequest)
    lc, _ := lambdacontext.FromContext(c.Context())

    return c.String(http.StatusOK, fmt.Sprintf("%s %s", e.RequestContext.AccountID, lc.AwsRequestID))
})
```

## Configuration
Handler configuration can be optionally specified by using `NewWithConfig`.

### Event Types
Rack supports API Gateway proxy integration, API Gateway V2 HTTP and ALB target group events. By default the event type is resolved at runtime, but this behaviour can be configured as required. The following example configures the handler to marshal to/from V2 HTTP events regardless of the payload.
```
cfg := rack.Config{
    Resolver:   rack.NewStaticResolver(rack.APIGatewayV2HTTPEventProcessor),
}

h := rack.NewWithConfig(cfg, handler)
```

### Middleware
Middleware can be specified by passing a `MiddlewareFunc` in the configuration. The `Chain` helper function allows multiple middleware functions to be combined into a single chain. Functions execute in the order they are specified as arguments.
```
cfg := rack.Config{
    Middleware: rack.Chain(errorLogging, extractClaims),
}

h := rack.NewWithConfig(cfg, handler)
```

### Error Handling
By default Rack will only return a function error if the incoming our outgoing payloads cannot be marshalled. All handler errors will be written to the response as a JSON body. This behaviour can be customised by modifying the handler `OnError` function. The following example writes the error message to the response as a string.
```
cfg := rack.Config{
    OnError: func(c rack.Context, err error) error {
        return c.String(rack.StatusCode(err), err.Error())
    },
}

h := rack.NewWithConfig(cfg, handler)
```

### Bind
The handler `Context` offers a `Bind` function to marshal the incoming JSON body into an object. It is possible to configure a post-bind operation, for example to perform validation.
```
cfg := rack.Config{
    OnBind: func(c rack.Context, v interface{}) error {
        if err := validate(v); err != nil {
            return rack.WrapError(http.StatusBadRequest, err)
        }

        return nil
    },
}

h := rack.NewWithConfig(cfg, func(c rack.Context) error {
    var t Task
    if err := c.Bind(&t); err != nil {
        return err
    }

    if err := store.CreateTask(c.Context(), t); err != nil {
        return err
    }

    return c.NoContent(http.StatusCreated)
})
```
