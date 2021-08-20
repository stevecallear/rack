package main

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/stevecallear/rack"
)

func main() {
	h := rack.New(func(c rack.Context) error {
		r := c.Request()
		return c.String(http.StatusOK, fmt.Sprintf("%s %s", r.Method, r.RawPath))
	})

	lambda.StartHandler(h)
}
