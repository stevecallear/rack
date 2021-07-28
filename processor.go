package rack

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/tidwall/gjson"
)

type (
	// Processor represents an event processor
	Processor interface {
		// CanProcess returns true if the processor is valid for the payload
		CanProcess(payload []byte) bool

		// UnmarshalRequest unmarshals the specified payload into a canonical request
		UnmarshalRequest(payload []byte) (*Request, error)

		// MarshalResponse marshals the canonical response into a response payload
		MarshalResponse(res *Response) ([]byte, error)
	}

	processor struct {
		canProcess       func([]byte) bool
		unmarshalRequest func([]byte) (*Request, error)
		marshalResponse  func(*Response) ([]byte, error)
	}
)

var (
	// APIGatewayProxyEventProcessor is an api gateway proxy event processor
	APIGatewayProxyEventProcessor Processor = &processor{
		canProcess: func(payload []byte) bool {
			pv := gjson.GetManyBytes(payload, "version", "requestContext.apiId")
			return !pv[0].Exists() && pv[1].Exists()
		},
		unmarshalRequest: func(payload []byte) (*Request, error) {
			e := new(events.APIGatewayProxyRequest)
			if err := json.Unmarshal(payload, e); err != nil {
				return nil, err
			}

			q := url.Values(e.MultiValueQueryStringParameters)
			h := http.Header(e.MultiValueHeaders)

			return &Request{
				Method:  e.HTTPMethod,
				RawPath: e.Path,
				Path:    e.PathParameters,
				Query:   q,
				Header:  h,
				Body:    e.Body,
				Event:   e,
			}, nil
		},
		marshalResponse: func(r *Response) ([]byte, error) {
			return json.Marshal(&events.APIGatewayProxyResponse{
				StatusCode:        r.StatusCode,
				Headers:           reduceHeaders(r.Headers),
				MultiValueHeaders: r.Headers,
				Body:              r.Body,
				IsBase64Encoded:   false,
			})
		},
	}

	// APIGatewayV2HTTPEventProcessor is an api gateway v2 http event processor
	APIGatewayV2HTTPEventProcessor Processor = &processor{
		canProcess: func(payload []byte) bool {
			pv := gjson.GetManyBytes(payload, "version", "requestContext.apiId")
			return pv[0].String() == "2.0" && pv[1].Exists()
		},
		unmarshalRequest: func(payload []byte) (*Request, error) {
			e := new(events.APIGatewayV2HTTPRequest)
			if err := json.Unmarshal(payload, e); err != nil {
				return nil, err
			}

			q := url.Values{}
			for k, ps := range e.QueryStringParameters {
				for _, v := range strings.Split(ps, ",") {
					q.Add(k, v)
				}
			}

			h := http.Header{}
			mergeMaps(e.Headers, nil, h.Add)

			return &Request{
				Method:  e.RequestContext.HTTP.Method,
				RawPath: e.RequestContext.HTTP.Path,
				Path:    e.PathParameters,
				Query:   q,
				Header:  h,
				Body:    e.Body,
				Event:   e,
			}, nil
		},
		marshalResponse: func(r *Response) ([]byte, error) {
			return json.Marshal(&events.APIGatewayV2HTTPResponse{
				StatusCode:        r.StatusCode,
				Headers:           reduceHeaders(r.Headers),
				MultiValueHeaders: r.Headers,
				Body:              r.Body,
				IsBase64Encoded:   false,
				Cookies:           []string{},
			})
		},
	}

	// ALBTargetGroupEventProcessor is an alb target group event processor
	ALBTargetGroupEventProcessor Processor = &processor{
		canProcess: func(payload []byte) bool {
			return gjson.GetBytes(payload, "requestContext.elb").Exists()
		},
		unmarshalRequest: func(payload []byte) (*Request, error) {
			e := new(events.ALBTargetGroupRequest)
			if err := json.Unmarshal(payload, e); err != nil {
				return nil, err
			}

			q := url.Values{}
			mergeMaps(e.QueryStringParameters, e.MultiValueQueryStringParameters, q.Add)

			h := http.Header{}
			mergeMaps(e.Headers, e.MultiValueHeaders, h.Add)

			return &Request{
				Method:  e.HTTPMethod,
				RawPath: e.Path,
				Path:    map[string]string{},
				Query:   q,
				Header:  h,
				Body:    e.Body,
				Event:   e,
			}, nil
		},
		marshalResponse: func(r *Response) ([]byte, error) {
			return json.Marshal(&events.ALBTargetGroupResponse{
				StatusCode:        r.StatusCode,
				StatusDescription: http.StatusText(r.StatusCode),
				Headers:           reduceHeaders(r.Headers),
				MultiValueHeaders: r.Headers,
				Body:              r.Body,
				IsBase64Encoded:   false,
			})
		},
	}
)

func (p *processor) CanProcess(payload []byte) bool {
	return p.canProcess(payload)
}

func (p *processor) UnmarshalRequest(payload []byte) (*Request, error) {
	return p.unmarshalRequest(payload)
}

func (p *processor) MarshalResponse(res *Response) ([]byte, error) {
	return p.marshalResponse(res)
}

func mergeMaps(sv map[string]string, mv map[string][]string, addFn func(k, v string)) {
	for k, v := range sv {
		addFn(k, v)
	}

	for k, vs := range mv {
		for _, v := range vs {
			addFn(k, v)
		}
	}
}

func reduceHeaders(h http.Header) map[string]string {
	m := make(map[string]string, len(h))
	for k := range h {
		m[k] = h.Get(k)
	}

	return m
}
