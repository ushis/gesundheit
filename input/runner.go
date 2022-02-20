package input

import "github.com/ushis/gesundheit/check"

type Runner struct {
	Input
}

func NewRunner(i Input) *Runner {
	return &Runner{i}
}

func (r *Runner) Run(events chan<- check.Event) {
	r.Input.Run(events)
}

func (r *Runner) Close() {
	r.Input.Close()
}
