package rack_test

import (
	"testing"

	"github.com/stevecallear/rack"
)

func TestNewStaticResolver(t *testing.T) {
	t.Run("should return the processor", func(t *testing.T) {
		exp := &testProcessor{canProcess: true}
		sut := rack.NewStaticResolver(exp)

		act, err := sut.Resolve(nil)
		assertErrorExists(t, err, false)
		if act != exp {
			t.Errorf("got %v, expecte %v", act, exp)
		}
	})
}

func TestNewDynamicResolver(t *testing.T) {
	proc := &testProcessor{canProcess: true}

	tests := []struct {
		name  string
		procs []rack.Processor
		exp   rack.Processor
		err   bool
	}{
		{
			name: "should return an error if there are no valid processors",
			procs: []rack.Processor{
				&testProcessor{canProcess: false},
			},
			err: true,
		},
		{
			name: "should return the first valid processor",
			procs: []rack.Processor{
				&testProcessor{canProcess: false},
				proc,
				&testProcessor{canProcess: true},
			},
			exp: proc,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := rack.NewConditionalResolver(tt.procs...)

			act, err := sut.Resolve(nil)
			assertErrorExists(t, err, tt.err)
			if act != tt.exp {
				t.Errorf("got %v, expected %v", act, tt.exp)
			}
		})
	}
}

type testProcessor struct {
	canProcess bool
}

func (p *testProcessor) CanProcess(payload []byte) bool {
	return p.canProcess
}

func (p *testProcessor) UnmarshalRequest([]byte) (*rack.Request, error) {
	panic("not implemented")
}

func (p *testProcessor) MarshalResponse(*rack.Response) ([]byte, error) {
	panic("not implemented")
}
