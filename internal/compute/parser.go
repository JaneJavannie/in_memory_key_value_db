package compute

import (
	"log/slog"
)

// Parser компонент внутри слоя, отвечающий за парсинг запроса
type Parser struct{}

func newParser() Parser { return Parser{} }

func (p *Parser) parse(text string) ([]string, error) {
	m := newStateMachine()

	for _, sym := range text {
		switch {

		case isArgumentSymbol(sym):
			_, err := m.currentWord.WriteRune(sym)
			if err != nil {
				return nil, err
			}
			m.processEvent(eventArgumentSymbol)

		case isWhiteSpace(sym):
			m.processEvent(eventSpace)

		default:
			slog.Error("unknown symbol '%c'", sym)
		}
	}

	return m.commandArgs, nil
}

func isArgumentSymbol(c rune) bool {
	return isDigit(c) || isLetter(c) || isPunctuation(c)
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}
func isLetter(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
}

func isPunctuation(r rune) bool {
	return r == '*' || r == '/' || r == '_'
}

func isWhiteSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n'
}
