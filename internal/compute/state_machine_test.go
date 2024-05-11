package compute

import (
	"log/slog"
	"testing"
)

func TestProcessEvent(t *testing.T) {
	// Create a test case with the sequence of events and the expected final state
	tests := []struct {
		name                string
		events              []event
		expectedState       state
		expectedCommandArgs []string
	}{
		{
			name:          " V ",
			events:        []event{eventSpace, eventArgumentSymbol, eventSpace},
			expectedState: stateSpace,
		},

		{
			name:          "2 V",
			events:        []event{eventArgumentSymbol, eventSpace, eventArgumentSymbol},
			expectedState: stateWord,
		},

		{
			name:          "3  ",
			events:        []event{eventArgumentSymbol, eventSpace, eventSpace},
			expectedState: stateSpace,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := newStateMachine(slog.Default())

			for _, event := range test.events {
				m.processEvent(event)
			}

			if m.state != test.expectedState {
				t.Errorf("For events %v, expected state %v, but got %v", test.events, test.expectedState, m.state)
			}
		})
	}
}
