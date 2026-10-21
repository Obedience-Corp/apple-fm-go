# apple-fm-go

`apple-fm-go` is an unofficial Go client for the `fm` command-line tool that ships with macOS 27 and runs Apple's on-device Foundation Model, shown in [WWDC26 session 334](https://developer.apple.com/videos/play/wwdc2026/334/). It is not affiliated with or endorsed by Apple. The subcommand and flag names it sends are provisional: they come from session 334 and early writeups of the macOS 27 CLI, and they have not yet been checked against `fm --help` on macOS 27.

The API is unstable until `v1.0.0`.

## Requirements

- macOS 27 on a Mac that supports Apple Intelligence, with Apple Intelligence turned on and the model downloaded.
- The `fm` terms accepted once by an operator with `sudo fm license`. This library never runs `sudo` and never accepts the terms for you.
- `fm available` succeeding in a terminal.

The package builds and its unit tests pass on any OS, including Linux, because it only executes a binary.

## Install

```sh
go get github.com/Obedience-Corp/apple-fm-go
```

## Quick start

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Obedience-Corp/apple-fm-go/pkg/fm"
)

func main() {
	ctx := context.Background()
	client := fm.New("fm") // an empty string also looks up "fm" on PATH

	if _, err := client.Available(ctx); err != nil {
		switch {
		case errors.Is(err, fm.ErrNotFound):
			log.Fatal("fm is not installed; it ships with macOS 27")
		case errors.Is(err, fm.ErrLicense):
			log.Fatal("run: sudo fm license")
		default:
			log.Fatal(err)
		}
	}

	resp, err := client.Respond(ctx, "What does chmod do?", &fm.RespondOptions{
		Instructions: "Answer in one sentence.",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(resp.Stdout)

	n, err := client.CountTokens(ctx, fm.CountRequest{Text: "A document to measure."})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("tokens:", n)

	schema, err := client.SchemaObject(ctx, fm.ObjectSchema{
		Name: "Person",
		Fields: []fm.Field{
			{Name: "name", Type: fm.String},
			{Name: "age", Type: fm.Int},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", schema) // write to a file and pass it as RespondOptions.SchemaPath
}
```

Every method takes a `context.Context`. Canceling it kills `fm` and any process it started.

`Client.Env` is the child environment. Leave it nil to pass through the current process environment.

## Respond options

| Field | Behavior |
| --- | --- |
| `Instructions` | Instruction text for the model. |
| `Model` | Leave empty for the CLI default, or pass `system`. Other names are passed through unchanged. `pcc` from session 334 is not supported until `fm --help` lists it. |
| `ImagePaths` | One image flag per path, in order. |
| `SchemaPath` | An existing schema file. The reply must be JSON and is returned in `Response.JSON`. Use `SchemaObject` to produce the file. |
| `Text` | A document passed with the text flag instead of a positional prompt. When `Text` is set, pass an empty prompt and put the task in `Instructions`. |

## Errors

Failures return an `*fm.Error` with the subcommand, exit code, stdout, and stderr. Match the category with `errors.Is`:

| Error | Meaning |
| --- | --- |
| `ErrNotFound` | `fm` is not on `PATH` or `Client.Bin` does not exist |
| `ErrLicense` | the `fm` terms are not accepted |
| `ErrUnavailable` | the model is not downloaded, or Apple Intelligence is off |
| `ErrContextWindow` | the prompt does not fit |
| `ErrGuardrail` | the model refused the prompt |
| `ErrUnknownModel` | `fm` rejected the model name |
| `ErrSchema` | `fm` rejected the schema, or a field cannot be expressed as flags |
| `ErrSchemaOutput` | the reply or schema was not JSON |
| `ErrTokenCount` | `fm count-tokens` did not print one integer |

Until exit codes from macOS 27 are captured, categories are detected from stable prefixes on stderr lines. The full stderr is always kept on the error.

## Provisional CLI mapping

| Go | `fm` arguments |
| --- | --- |
| `Available` | `available` |
| `Respond` | `respond [--instructions=…] [--model=…] [--image=…]… [--schema=…] (--text=… \| -- <prompt>)` |
| `CountTokens` | `count-tokens --text=…` |
| `SchemaObject` | `schema object --name=… --string=… --int=… [--array]` |

All of these strings live in `pkg/fm/flags.go`, so a correction after the macOS 27 help capture is a one-file change.

## Examples

`examples/respond`, `examples/schema`, and `examples/image` are small programs that need a real `fm`:

```sh
go run ./examples/respond "What does chmod do?"
go run ./examples/schema
go run ./examples/image Screenshot.png
```

## Development

```sh
just test          # unit tests with a fake fm; no Apple tooling needed
just lint          # golangci-lint, or go vet when it is not installed
just test-applefm  # runs the real fm available on macOS 27; skips when fm is missing
```

## License

Apache-2.0. See [LICENSE](LICENSE).
