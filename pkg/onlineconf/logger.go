package onlineconf

import (
	"context"
	"log"
)

type DefaultLogger struct{}

func (l *DefaultLogger) Warn(ctx context.Context, msg string, args ...any) {
	log.Printf("[WARN] %s: %+v\n", msg, args)
}

func (l *DefaultLogger) Error(ctx context.Context, msg string, args ...any) {
	log.Printf("[ERRO] %s: %+v\n", msg, args)
}

func (l *DefaultLogger) Fatal(ctx context.Context, msg string, args ...any) {
	log.Fatalf("[FATA] %s: %+v\n", msg, args)
}
