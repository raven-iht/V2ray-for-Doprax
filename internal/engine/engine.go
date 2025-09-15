package engine

type Input map[string]any

type Output map[string]any

type DecisionFunc func(Input) (Output, error)

type Engine struct {
	models map[string]map[string]DecisionFunc
}

func New() *Engine {
	return &Engine{models: make(map[string]map[string]DecisionFunc)}
}

func (e *Engine) Register(model, decision string, fn DecisionFunc) {
	if e.models[model] == nil {
		e.models[model] = make(map[string]DecisionFunc)
	}
	e.models[model][decision] = fn
}

func (e *Engine) Evaluate(model, decision string, in Input) (Output, bool, error) {
	m, ok := e.models[model]
	if !ok { return nil, false, nil }
	fn, ok := m[decision]
	if !ok { return nil, false, nil }
	out, err := fn(in)
	return out, true, err
}
