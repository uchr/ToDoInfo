package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/uchr/ToDoInfo/internal/cli"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := cli.Execute(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}