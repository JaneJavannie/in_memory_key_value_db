package compute

import "context"

// compute - слой, отвечающий за обработку запроса

type parser interface {
	parse(string) ([]string, error)
}

type analyzer interface {
	analyzeQuery(ctx context.Context, parsed []string) (Query, error)
	validate(ctx context.Context, parsed []string) error
}

type Computer struct{}

func NewComputer() Computer { return Computer{} }

func (c *Computer) Compute(ctx context.Context, text string) (Query, error) {
	p := newParser()

	result, err := p.parse(text)
	if err != nil {
		return Query{}, err
	}

	a := newAnalyzer()

	query, err := a.analyzeQuery(ctx, result)
	if err != nil {
		return Query{}, err
	}

	return query, nil
}
