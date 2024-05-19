package consts

import "errors"

const RequestID = "request_id"

const (
	CommandSet = "SET"
	CommandGet = "GET"
	CommandDel = "DEL"
)

var (
	ErrParseSymbol = errors.New("parse error")

	ErrInvalidSetQueryArgs = errors.New("invalid set query args")
	ErrInvalidGetQueryArgs = errors.New("invalid get query args")
	ErrInvalidDelQueryArgs = errors.New("invalid del query args")
)
