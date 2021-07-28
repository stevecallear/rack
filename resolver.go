package rack

import "errors"

type (
	// Resolver represents an event processor resolver
	Resolver interface {
		Resolve(payload []byte) (Processor, error)
	}

	resolverFunc func([]byte) (Processor, error)
)

var (
	// ErrUnsupportedEventType indicates that the supplied event payload is not supported
	ErrUnsupportedEventType = errors.New("unsupported event type")

	defaultResolver = NewConditionalResolver(
		APIGatewayProxyEventProcessor,
		APIGatewayV2HTTPEventProcessor,
		ALBTargetGroupEventProcessor,
	)
)

// NewStaticResolver returns a new static event processor resolver
// The supplied processor will be invoked for marshal/unmarshal
// operations, regardless of the incoming payload.
func NewStaticResolver(p Processor) Resolver {
	return resolverFunc(func([]byte) (Processor, error) {
		return p, nil
	})
}

// NewConditionalResolver returns a new conditional event processor resolver
// The first applicable processor will be returned, based on the
// incoming payload.
func NewConditionalResolver(p ...Processor) Resolver {
	return resolverFunc(func(payload []byte) (Processor, error) {
		for _, pp := range p {
			if pp.CanProcess(payload) {
				return pp, nil
			}
		}

		return nil, ErrUnsupportedEventType
	})
}

// Resolve resolves a resolver for the specified payload
func (r resolverFunc) Resolve(payload []byte) (Processor, error) {
	return r(payload)
}
