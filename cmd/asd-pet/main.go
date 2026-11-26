// asd-pet is the opt-in visual demo. Double-clicking it opens only the pet;
// it cannot fall through to the terminal manager or initialize personal state.
package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/liuhui-code/agent-session-desk/internal/cli"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	args := append([]string{"pet-demo"}, os.Args[1:]...)
	code := cli.Run(ctx, args, os.Stdin, os.Stdout, os.Stderr)
	cancel()
	os.Exit(code)
}
