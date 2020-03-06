package main

import (
	"context"

	"github.com/m-mizutani/deepalert"
	"github.com/m-mizutani/deepalert/emitter"
)

func handler(ctx context.Context, report deepalert.Report) error {
	return nil
}

func main() {
	emitter.Start(handler)
}
