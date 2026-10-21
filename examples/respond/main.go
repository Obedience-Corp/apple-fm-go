package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/Obedience-Corp/apple-fm-go/pkg/fm"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	prompt := strings.Join(os.Args[1:], " ")
	if prompt == "" {
		prompt = "What does chmod do?"
	}

	client := fm.New("")
	resp, err := client.Respond(ctx, prompt, &fm.RespondOptions{
		Instructions: "Answer in one sentence.",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(resp.Stdout)
}
