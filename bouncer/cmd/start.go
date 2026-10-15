package cmd

import (
	"context"
	"fmt"
	"log/slog"

	pg "github.com/divyanshu-parihar/systems/bouncer/pg"
)

const dsn = "postgresql://postgres:JkzEoDRXKdhyovUuUohsoXiQwhCbZXoK@maglev.proxy.rlwy.net:15058/railway"

func Run() error {
	fmt.Println("Starting the bouncer...")
	ctx := context.Background()
	bouncer, err := pg.NewPostgresBouncer(dsn, 10, 2)
	if err != nil {
		slog.Error("Error creating bouncer: ", "err", err)
		return err
	}

	bouncer.AddConnection(ctx, 0, dsn)
	bouncer.AddConnection(ctx, 1, dsn)
	bouncer.Serve(ctx)
	return nil
}
