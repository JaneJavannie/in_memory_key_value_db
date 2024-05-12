package compute

import (
	"fmt"
	"log/slog"

	"github.com/JaneJavannie/in_memory_key_value_db/internal/consts"
)

// Parser компонент внутри слоя, отвечающий за парсинг запроса
type Parser struct {
	logger *slog.Logger
}

func newParser(logger *slog.Logger) Parser {
	return Parser{
		logger: logger,
	}
}

func (p *Parser) parse(text string) ([]string, error) {
	m := newStateMachine(p.logger)

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
			return nil, fmt.Errorf("%w: unknown symbol: %q", consts.ErrParseSymbol, sym)
		}
	}

	if m.currentWord.Len() > 0 {
		m.commandArgs = append(m.commandArgs, m.currentWord.String())
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
