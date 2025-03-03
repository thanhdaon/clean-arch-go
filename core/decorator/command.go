package decorator

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

func ApplyCommandDecorators[H any](handler CommandHandler[H], logger *logrus.Entry) CommandHandler[H] {
	return commandLoggingDecorator[H]{
		base: commandTracingDecorator[H]{
			base: handler,
		},
		logger: logger,
	}
}

type CommandHandler[C any] interface {
	Handle(ctx context.Context, cmd C) error
}

func generateActionName(handler any) string {
	return strings.Split(fmt.Sprintf("%T", handler), ".")[1]
}
