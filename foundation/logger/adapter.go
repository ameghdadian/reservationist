package logger

import (
	"context"
	"os"
)

type AsyncLogAdapter struct {
	baseContext context.Context
	log         *Logger
}

func NewAsyncLogAdapter(ctx context.Context, log *Logger) *AsyncLogAdapter {
	return &AsyncLogAdapter{
		baseContext: ctx,
		log:         log,
	}
}

func (adp *AsyncLogAdapter) Debug(args ...any) {
	args = append(args, "msg")
	adp.log.Debug(adp.baseContext, "LOGGER", args...)
}

func (adp *AsyncLogAdapter) Info(args ...any) {
	args = append(args, "msg")
	adp.log.Info(adp.baseContext, "LOGGER", args...)
}

func (adp *AsyncLogAdapter) Warn(args ...any) {
	args = append(args, "msg")
	adp.log.Warn(adp.baseContext, "LOGGER", args...)
}

func (adp *AsyncLogAdapter) Error(args ...any) {
	args = append(args, "msg")
	adp.log.Error(adp.baseContext, "LOGGER", args...)
}

func (adp *AsyncLogAdapter) Fatal(args ...any) {
	args = append(args, "msg")
	adp.log.Error(adp.baseContext, "LOGGER", args...)
	os.Exit(1)
}
