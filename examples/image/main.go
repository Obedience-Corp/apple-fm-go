package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/Obedience-Corp/apple-fm-go/pkg/fm"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: image <path-to-image>")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	client := fm.New("")
	resp, err := client.Respond(ctx, "What app is the user using in this screenshot?", &fm.RespondOptions{
		ImagePaths: []string{os.Args[1]},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(resp.Stdout)
}
