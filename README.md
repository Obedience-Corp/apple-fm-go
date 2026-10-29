# apple-fm-go

`apple-fm-go` is an unofficial Go client for Apple's `fm` command-line tool. It provides context-aware calls for text responses, image prompts, structured output, availability checks, and token counting. It executes a subprocess and has no CGo, Swift, or Python dependency.

**Status: experimental, unreleased.** The API and CLI mapping are provisional. Fake-process tests verify the Go wrapper; a real macOS 27 `fm` compatibility run is still required before `v0.1.0`. This project is not affiliated with or endorsed by Apple.

## Requirements and compatibility

- Go 1.24 or later.
- To run inference: macOS 27 and an Apple Intelligence-capable Mac with the model available. Apple introduced the CLI in [WWDC26 session 334](https://developer.apple.com/videos/play/wwdc2026/334/).
- Follow the installed CLI's setup and license instructions. The library never runs `sudo` or accepts terms. The earlier `sudo fm license` setup assumption remains unverified.

The module and examples compile for Linux, macOS, and Windows on amd64 and arm64. Compilation does not imply that Apple's executable is available on those platforms: callers can import the package on unsupported hosts and handle `ErrNotFound` when invoking it.

This is a CLI wrapper, not a binding to the complete Foundation Models framework. Apple's separate [Python SDK](https://github.com/apple/python-apple-fm-sdk) supports macOS 26.0+ with Xcode 26.0+ and Python 3.10+; that does **not** make the `fm` executable available on macOS 26.

## Apple API coverage

Reviewed against Apple's published documentation on **2026-10-29**.

| Capability | This Go SDK |
| --- | --- |
| One-shot text, instructions, image paths | `Respond`; buffers output until the process exits |
| Model selection | `RespondOptions.Model` passes the value through; Apple demonstrates `pcc` for Private Cloud Compute |
| Guided output | `SchemaObject` builds a schema; `SchemaPath` requests JSON; string and integer fields only |
| Availability and token counting | `Available`, `CountTokens`; command/output contracts still await native verification |
| Streaming, conversation history, chat save/resume | Not exposed; each call starts a fresh process |
| Tools, Dynamic Profiles, custom providers, reasoning controls | Not exposed by this wrapper |

Apple's [Foundation Models updates](https://developer.apple.com/documentation/updates/foundationmodels) describe newer model APIs, image analysis, Dynamic Profiles, tool-calling controls, and revised Swift error types. These framework APIs do not establish equivalent CLI flags or exit codes. See the [compatibility review](docs/compatibility.md) for evidence and outstanding checks.

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
	"time"

	"github.com/Obedience-Corp/apple-fm-go/pkg/fm"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	client := fm.New("fm") // an empty string also looks up "fm" on PATH

	if _, err := client.Available(ctx); err != nil {
		switch {
		case errors.Is(err, fm.ErrNotFound):
			log.Fatal("fm is not installed; it ships with macOS 27")
		case errors.Is(err, fm.ErrLicense):
			log.Fatal("review the installed fm license/setup instructions")
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

Every method takes a `context.Context`. Canceling it kills `fm`. On Unix the whole process group dies, so helpers `fm` started die too; on Windows only `fm` itself is killed.

`Client.Env` is the child environment. Leave it nil to pass through the current process environment.

## Respond options

| Field | Behavior |
| --- | --- |
| `Instructions` | Instruction text for the model. |
| `Model` | Leave empty for the CLI default. Apple documents `pcc` for Private Cloud Compute (subject to usage limits). Values pass through unchanged; `system` has not been confirmed as an explicit CLI model name. |
| `ImagePaths` | One image flag per path, in order. |
| `SchemaPath` | An existing schema file. The reply must be JSON and is returned in `Response.JSON`. Use `SchemaObject` to produce the file. |
| `Text` | A document passed with the text flag instead of a positional prompt. When `Text` is set, pass an empty prompt and put the task in `Instructions`. |

## Errors

Subprocess failures return an `*fm.Error` with the subcommand, exit code, stdout, and stderr. Local argument validation may return other errors. Cancellation remains discoverable through `errors.Is(err, context.Canceled)` or `context.DeadlineExceeded`. Match subprocess categories with `errors.Is`:

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

Until exit codes from macOS 27 are captured, categories are best-effort matches against assumed stderr prefixes, not an Apple error-code contract. Unknown failures remain unclassified; the full stderr is always kept on the error.

## Provisional CLI mapping

| Go | `fm` arguments |
| --- | --- |
| `Available` | `available` |
| `Respond` | `respond [--instructions=…] [--model=…] [--image=…]… [--schema=…] (--text=… \| -- <prompt>)` |
| `CountTokens` | `count-tokens --text=…` |
| `SchemaObject` | `schema object --name=… --string=… --int=… [--array]` |

Flag names live in `pkg/fm/flags.go`; argument ordering and output parsing live in the corresponding operation files. The [compatibility review](docs/compatibility.md) separates Apple-documented spellings from assumptions. In particular, `available`, `count-tokens`, `--text`, `--int`, and error prefixes still require native confirmation.

## Examples

`examples/respond`, `examples/schema`, and `examples/image` are small programs that need a real `fm`:

```sh
go run ./examples/respond "What does chmod do?"
go run ./examples/schema
go run ./examples/image Screenshot.png
```

## Development

```sh
just test                  # fake-process tests with race detection, isolated in Docker
just lint                  # formatting check and go vet, including the applefm tag
just check cross-build     # Linux/macOS/Windows, amd64/arm64; no CGo
just test-applefm           # native smoke tests only; skips if fm/model is unavailable
just check native-required # same native tests, but missing prerequisites fail
```

The unit suite creates subprocess fixtures and runs inside a disposable Linux container with read-only source and networking disabled. Docker and the `golang:1.24-bookworm` image are required. Recipes are grouped in `.justfiles/check.just`.

The native suite checks availability, a nonempty on-device response, schema JSON, and token counts with bounded contexts. It does not opt into PCC or upload images. A skipped native suite is not compatibility evidence; use `native-required` for release validation. Image/PCC behavior and schema round-trips still need separate native verification.

## Related

- [obey-go-sdk](https://github.com/Obedience-Corp/obey-go-sdk) — Go client for the Obey daemon.
- [claude-code-go](https://github.com/lancekrogers/claude-code-go) — Claude Code CLI SDK.
- [muse-spark-go](https://github.com/Obedience-Corp/muse-spark-go) — Muse CLI and Session Protocol SDK.

## License

Apache-2.0. See [LICENSE](LICENSE).
