package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/Obedience-Corp/apple-fm-go/pkg/fm"
)

type person struct {
	Name     string   `json:"name"`
	Age      int      `json:"age"`
	Projects []string `json:"projects"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	client := fm.New("")
	schema, err := client.SchemaObject(ctx, fm.ObjectSchema{
		Name: "Person",
		Fields: []fm.Field{
			{Name: "name", Type: fm.String},
			{Name: "age", Type: fm.Int},
			{Name: "projects", Type: fm.String, Repeated: true},
		},
	})
	if err != nil {
		return fmt.Errorf("build schema: %w", err)
	}

	dir, err := os.MkdirTemp("", "apple-fm-go-schema-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	schemaPath := filepath.Join(dir, "person.json")
	if err := os.WriteFile(schemaPath, schema, 0o600); err != nil {
		return err
	}

	resp, err := client.Respond(ctx, "", &fm.RespondOptions{
		Instructions: "Extract the person described in the text.",
		Text:         "Ada is 36 and maintains the parser and the build scripts.",
		SchemaPath:   schemaPath,
	})
	if err != nil {
		return fmt.Errorf("respond: %w", err)
	}

	var p person
	if err := json.Unmarshal(resp.JSON, &p); err != nil {
		return fmt.Errorf("decode reply: %w", err)
	}
	fmt.Printf("%+v\n", p)
	return nil
}
