package rack_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/aws/aws-lambda-go/events"

	"github.com/stevecallear/rack"
)

func TestAPIGatewayProxyEventProcessor_CanProcess(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		exp     bool
	}{
		{
			name:    "should return true for api gateway proxy events",
			payload: []byte(apiGatewayProxyEventPayload),
			exp:     true,
		},
		{
			name:    "should return false for api gateway v2 http events",
			payload: []byte(apiGatewayV2HTTPEventPayload),
			exp:     false,
		},
		{
			name:    "should return false for alb target group events",
			payload: []byte(albTargetGroupSingleValueEventPayload),
			exp:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := rack.APIGatewayProxyEventProcessor
			act := sut.CanProcess(tt.payload)

			if act != tt.exp {
				t.Errorf("got %v, expected %v", act, tt.exp)
			}
		})
	}
}

func TestAPIGatewayProxyEventProcessor_UnmarshalRequest(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		exp     *rack.Request
		err     bool
	}{
		{
			name:    "should return an error if the payload is invalid",
			payload: []byte("{"),
			err:     true,
		},
		{
			name:    "should return the request",
			payload: []byte(apiGatewayProxyEventPayload),
			exp: &rack.Request{
				Method:  http.MethodGet,
				RawPath: "/resource/",
				Path: map[string]string{
					"proxy": "resource",
				},
				Query: url.Values{
					"q1": {"v1"},
					"q2": {"v2", "v3"},
				},
				Header: http.Header{
					"X-Custom-Header1": {"v1"},
					"X-Custom-Header2": {"v2", "v3"},
				},
				Body:  "body",
				Event: unmarshal([]byte(apiGatewayProxyEventPayload), new(events.APIGatewayProxyRequest)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := rack.APIGatewayProxyEventProcessor
			act, err := sut.UnmarshalRequest(tt.payload)
			assertErrorExists(t, err, tt.err)
			assertDeepEqual(t, act, tt.exp)
		})
	}
}

func TestAPIGatewayProxyEventProcessor_MarshalResponse(t *testing.T) {
	t.Run("should marshal the response", func(t *testing.T) {
		res := &rack.Response{
			StatusCode: http.StatusOK,
			Headers: http.Header{
				"X-Custom-Header1": {"v1"},
				"X-Custom-Header2": {"v2", "v3"},
			},
			Body: "body",
		}

		exp := marshal(&events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"X-Custom-Header1": "v1",
				"X-Custom-Header2": "v2",
			},
			MultiValueHeaders: map[string][]string{
				"X-Custom-Header1": {"v1"},
				"X-Custom-Header2": {"v2", "v3"},
			},
			Body: "body",
		})

		sut := rack.APIGatewayProxyEventProcessor
		act, err := sut.MarshalResponse(res)
		assertErrorExists(t, err, false)
		assertDeepEqual(t, act, exp)
	})
}

func TestAPIGatewayV2HTTPEventProcessor_CanProcess(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		exp     bool
	}{
		{
			name:    "should return true for api gateway v2 http events",
			payload: []byte(apiGatewayV2HTTPEventPayload),
			exp:     true,
		},
		{
			name:    "should return false for api gateway proxy events",
			payload: []byte(apiGatewayProxyEventPayload),
			exp:     false,
		},
		{
			name:    "should return false for alb target group events",
			payload: []byte(albTargetGroupSingleValueEventPayload),
			exp:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := rack.APIGatewayV2HTTPEventProcessor
			act := sut.CanProcess(tt.payload)

			if act != tt.exp {
				t.Errorf("got %v, expected %v", act, tt.exp)
			}
		})
	}
}

func TestAPIGatewayV2HTTPEventProcessor_UnmarshalRequest(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		exp     *rack.Request
		err     bool
	}{
		{
			name:    "should return an error if the payload is invalid",
			payload: []byte("{"),
			err:     true,
		},
		{
			name:    "should return the response",
			payload: []byte(apiGatewayV2HTTPEventPayload),
			exp: &rack.Request{
				Method:  http.MethodGet,
				RawPath: "/resource/",
				Path: map[string]string{
					"p": "v",
				},
				Query: url.Values{
					"q1": {"v1"},
					"q2": {"v2", "v3"},
				},
				Header: http.Header{
					"X-Custom-Header1": {"v1"},
					"X-Custom-Header2": {"v2"},
				},
				Body:  "body",
				Event: unmarshal([]byte(apiGatewayV2HTTPEventPayload), new(events.APIGatewayV2HTTPRequest)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := rack.APIGatewayV2HTTPEventProcessor
			act, err := sut.UnmarshalRequest(tt.payload)
			assertErrorExists(t, err, tt.err)
			assertDeepEqual(t, act, tt.exp)
		})
	}
}

func TestAPIGatewayV2HTTPEventProcessor_MarshalResponse(t *testing.T) {
	t.Run("should marshal the response", func(t *testing.T) {
		res := &rack.Response{
			StatusCode: http.StatusOK,
			Headers: http.Header{
				"X-Custom-Header1": {"v1"},
				"X-Custom-Header2": {"v2", "v3"},
			},
			Body: "body",
		}

		exp := marshal(&events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"X-Custom-Header1": "v1",
				"X-Custom-Header2": "v2",
			},
			MultiValueHeaders: map[string][]string{
				"X-Custom-Header1": {"v1"},
				"X-Custom-Header2": {"v2", "v3"},
			},
			Body:    "body",
			Cookies: []string{},
		})

		sut := rack.APIGatewayV2HTTPEventProcessor
		act, err := sut.MarshalResponse(res)
		assertErrorExists(t, err, false)
		assertDeepEqual(t, act, exp)
	})
}

func TestALBTargetGroupEventProcessor_CanProcess(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		exp     bool
	}{
		{
			name:    "should return true for alb target group events",
			payload: []byte(albTargetGroupSingleValueEventPayload),
			exp:     true,
		},

		{
			name:    "should return false for api gateway proxy events",
			payload: []byte(apiGatewayProxyEventPayload),
			exp:     false,
		},
		{
			name:    "should return false for api gateway v2 http events",
			payload: []byte(apiGatewayV2HTTPEventPayload),
			exp:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := rack.ALBTargetGroupEventProcessor
			act := sut.CanProcess(tt.payload)

			if act != tt.exp {
				t.Errorf("got %v, expected %v", act, tt.exp)
			}
		})
	}
}

func TestALBTargetGroupEventProcessor_UnmarshalRequest(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		exp     *rack.Request
		err     bool
	}{
		{
			name:    "should return an error if the payload is invalid",
			payload: []byte("{"),
			err:     true,
		},
		{
			name:    "should return the response for single value payloads",
			payload: []byte(albTargetGroupSingleValueEventPayload),
			exp: &rack.Request{
				Method:  http.MethodGet,
				RawPath: "/resource/",
				Path:    map[string]string{},
				Query: url.Values{
					"q1": {"v1"},
					"q2": {"v2"},
				},
				Header: http.Header{
					"X-Custom-Header1": {"v1"},
					"X-Custom-Header2": {"v2"},
				},
				Body:  "body",
				Event: unmarshal([]byte(albTargetGroupSingleValueEventPayload), new(events.ALBTargetGroupRequest)),
			},
		},
		{
			name:    "should return the response for multi value payloads",
			payload: []byte(albTargetGroupMultiValueEventPayload),
			exp: &rack.Request{
				Method:  http.MethodGet,
				RawPath: "/resource/",
				Path:    map[string]string{},
				Query: url.Values{
					"q1": {"v1"},
					"q2": {"v2", "v3"},
				},
				Header: http.Header{
					"X-Custom-Header1": {"v1"},
					"X-Custom-Header2": {"v2", "v3"},
				},
				Body:  "body",
				Event: unmarshal([]byte(albTargetGroupMultiValueEventPayload), new(events.ALBTargetGroupRequest)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := rack.ALBTargetGroupEventProcessor
			act, err := sut.UnmarshalRequest(tt.payload)
			assertErrorExists(t, err, tt.err)
			assertDeepEqual(t, act, tt.exp)
		})
	}
}

func TestALBTargetGroupEventProcessor_MarshalResponse(t *testing.T) {
	t.Run("should marshal the response", func(t *testing.T) {
		res := &rack.Response{
			StatusCode: http.StatusOK,
			Headers: http.Header{
				"X-Custom-Header1": {"v1"},
				"X-Custom-Header2": {"v2", "v3"},
			},
			Body: "body",
		}

		exp := marshal(&events.ALBTargetGroupResponse{
			StatusCode:        http.StatusOK,
			StatusDescription: http.StatusText(http.StatusOK),
			Headers: map[string]string{
				"X-Custom-Header1": "v1",
				"X-Custom-Header2": "v2",
			},
			MultiValueHeaders: map[string][]string{
				"X-Custom-Header1": {"v1"},
				"X-Custom-Header2": {"v2", "v3"},
			},
			Body: "body",
		})

		sut := rack.ALBTargetGroupEventProcessor
		act, err := sut.MarshalResponse(res)
		assertErrorExists(t, err, false)
		assertDeepEqual(t, act, exp)
	})
}

const (
	apiGatewayProxyEventPayload = `{
	"resource": "/{proxy+}",
	"path": "/resource/",
	"httpMethod": "GET",
	"headers": {
		"X-Custom-Header1": "v1",
		"X-Custom-Header2": "v3"
	},
	"multiValueHeaders": {
		"X-Custom-Header1": [
			"v1"
		],
		"X-Custom-Header2": [
			"v2",
			"v3"
		]
	},
	"queryStringParameters": {
		"q1": "v1",
		"q2": "v3"
	},
	"multiValueQueryStringParameters": {
		"q1": [
			"v1"
		],
		"q2": [
			"v2",
			"v3"
		]
	},
	"pathParameters": {
		"proxy": "resource"
	},
	"stageVariables": null,
	"requestContext": {
		"resourcePath": "/{proxy+}",
		"httpMethod": "GET",
		"path": "/dev/resource/",
		"protocol": "HTTP/1.1",
		"apiId": "apiid"
	},
	"body": "body",
	"isBase64Encoded": false
}`

	apiGatewayV2HTTPEventPayload = ` {
	"version": "2.0",
	"routeKey": "$default",
	"rawPath": "/resource/",
	"rawQueryString": "q1=v1&q2=v2&q2=v3",
	"pathParameters": {
		"p": "v"
	},
	"headers": {
		"x-custom-header1": "v1",
		"x-custom-header2": "v2"
	},
	"queryStringParameters": {
		"q1": "v1",
		"q2": "v2,v3"
	},
	"requestContext": {
		"apiId": "apiid",
		"http": {
			"method": "GET",
			"path": "/resource/",
			"protocol": "HTTP/1.1"
		}
	},
	"body": "body",
	"isBase64Encoded": false
}`

	albTargetGroupSingleValueEventPayload = `{
	"requestContext": {
		"elb": {
			"targetGroupArn": "arn"
		}
	},
	"httpMethod": "GET",
	"path": "/resource/",
	"queryStringParameters": {
		"q1": "v1",
		"q2": "v2"
	},
	"headers": {
		"x-custom-header1": "v1",
		"x-custom-header2": "v2"
	},
	"body": "body",
	"isBase64Encoded": false
}`

	albTargetGroupMultiValueEventPayload = `{
	"requestContext": {
		"elb": {
			"targetGroupArn": "arn"
		}
	},
	"httpMethod": "GET",
	"path": "/resource/",
	"multiValueQueryStringParameters": {
		"q1": [
			"v1"
		],
		"q2": [
			"v2",
			"v3"
		]
	},
	"multiValueHeaders": {
		"x-custom-header1": [
			"v1"
		],
		"x-custom-header2": [
			"v2",
			"v3"
		]
	},
	"body": "body",
	"isBase64Encoded": false
}`
)
