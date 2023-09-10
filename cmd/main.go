package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"giclo/internal/application"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt, syscall.SIGSEGV)
	defer cancel()
	go application.Start(ctx)
	<-ctx.Done()
	application.Stop(ctx)
}
