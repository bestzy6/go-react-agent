package tools

import "context"

type Tool interface {
	Name() string
	Description() string
	ArgsDescription() string
	Exec(ctx context.Context, input string) (string, error)
}
