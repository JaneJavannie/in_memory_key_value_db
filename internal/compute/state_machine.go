package compute

import (
	"log/slog"
	"strings"
)

type state string

var (
	stateInitial state = "initial"
	stateWord    state = "word"
	stateSpace   state = "white_space"
)

type event int

var (
	eventSpace          event = 0
	eventArgumentSymbol event = 1
)

type action string

var (
	actionWordEnd action = "word_end"
)

var transitionTable = map[state]map[event]finalState{}

type finalState struct {
	state  state
	action action
}

func initTransitionTable() {
	transitionTable[stateInitial] = map[event]finalState{}
	transitionTable[stateWord] = map[event]finalState{}
	transitionTable[stateSpace] = map[event]finalState{}

	transitionTable[stateInitial][eventArgumentSymbol] = finalState{state: stateWord} // ActionWordBegin
	transitionTable[stateInitial][eventSpace] = finalState{state: stateSpace}

	transitionTable[stateWord][eventArgumentSymbol] = finalState{state: stateWord}
	transitionTable[stateWord][eventSpace] = finalState{state: stateSpace, action: actionWordEnd}

	transitionTable[stateSpace][eventSpace] = finalState{state: stateSpace}
	transitionTable[stateSpace][eventArgumentSymbol] = finalState{state: stateWord} // ActionWordBegin
}

type stateMachine struct {
	logger      *slog.Logger
	state       state
	commandArgs []string
	currentWord strings.Builder
}

func newStateMachine(logger *slog.Logger) *stateMachine {
	initTransitionTable()
	return &stateMachine{
		logger:      logger,
		state:       stateInitial,
		commandArgs: make([]string, 0),
		currentWord: strings.Builder{},
	}
}

func (m *stateMachine) processEvent(event event) {
	finState, ok := transitionTable[m.state][event]
	if !ok {
		return
	}

	m.state = finState.state

	if finState.action == actionWordEnd {
		m.commandArgs = append(m.commandArgs, m.currentWord.String())
		m.currentWord.Reset()
	}
}
