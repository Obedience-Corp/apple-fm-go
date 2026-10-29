<h1 align="center">Apple Foundation Models for Go</h1>

<p align="center">
  A lightweight Go SDK for Apple's <code>fm</code> command-line tool.
</p>

<p align="center">
  <a href="#project-status"><img src="https://img.shields.io/badge/status-experimental-orange.svg" alt="Status: experimental"></a>
  <a href="go.mod"><img src="https://img.shields.io/badge/Go-1.24%2B-00ADD8.svg" alt="Go 1.24+"></a>
  <a href="#requirements"><img src="https://img.shields.io/badge/inference-macOS%2027-555555.svg" alt="Inference: macOS 27"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License: Apache 2.0"></a>
  <a href="https://github.com/Obedience-Corp/festival"><img src="https://img.shields.io/badge/built%20with-Festival-7c3aed.svg" alt="Built with Festival"></a>
</p>

<p align="center">
  <a href="#installation">Install</a> ·
  <a href="#quick-start">Quick start</a> ·
  <a href="#examples">Examples</a> ·
  <a href="docs/compatibility.md">Compatibility</a> ·
  <a href="#development">Development</a>
</p>

`apple-fm-go` brings Apple's `fm` CLI into Go applications: send prompts, attach images, request structured JSON, and propagate cancellation through the subprocess. The package uses only the Go standard library, with no CGo, Swift, or Python build dependency.

## Project status

**Experimental and unreleased.** The Go wrapper is tested with fake subprocesses; its CLI contract still needs verification against a real macOS 27 installation before `v0.1.0`. The [compatibility review](docs/compatibility.md) records the evidence and remaining checks. This is an independent project, not affiliated with or endorsed by Apple.

## Features

- **Text and image prompts** — instructions, model selection, and image paths in one request.
- **Structured output** — schema generation and JSON replies, with raw output retained on parse errors.
- **Context cancellation** — deadlines and cancellation reach every subprocess operation.
- **Inspectable errors** — match categories with `errors.Is` and inspect exit status, stdout, and stderr.
- **Portable imports** — compile the SDK and examples for Linux, macOS, and Windows on amd64 and arm64; inference requires Apple's executable.

Responses are buffered until the process exits. Streaming, persistent conversations, and tool calling are not exposed by this SDK.

## Requirements

| Requirement | Details |
| --- | --- |
| Build | Go 1.24 or later |
| Inference | macOS 27 on an Apple Intelligence-capable Mac with the model available |
| Executable | `fm` on `PATH`, or an explicit executable path passed to `fm.New` |
| Setup | Follow the installed CLI's setup and license instructions; the SDK never runs `sudo` or accepts terms |

Apple introduced the CLI in [WWDC26 session 334](https://developer.apple.com/videos/play/wwdc2026/334/). Apple's separate [Python SDK](https://github.com/apple/python-apple-fm-sdk) supports macOS 26.0+ with Xcode 26.0+ and Python 3.10+; those requirements do not apply to this CLI wrapper.

On a host without `fm`, the Go package still compiles and calls return `fm.ErrNotFound`. This allows applications to offer the provider conditionally.

## Installation

The repository is currently private. With repository access and Git authentication configured:

```sh
GOPRIVATE=github.com/Obedience-Corp/apple-fm-go \
  go get github.com/Obedience-Corp/apple-fm-go@main
```

There is no tagged release yet. Go records the resolved commit as a pseudo-version in your application's `go.mod`.

## Quick start

Send one prompt with a deadline:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/Obedience-Corp/apple-fm-go/pkg/fm"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()

    client := fm.New("") // Locate fm on PATH.
    response, err := client.Respond(ctx, "What does chmod do?", &fm.RespondOptions{
        Instructions: "Answer in one sentence.",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Print(response.Stdout)
}
```

Use `fm.New("/path/to/fm")` to choose an executable. Leave `Client.Env` nil to inherit the parent environment, or set it to supply the child's environment explicitly.

Every operation accepts a `context.Context`. Cancellation kills the process group on Unix, including helpers started by `fm`; on Windows it kills the direct subprocess.

## Examples

Run these programs from a checkout with a real `fm` installation:

| Example | Demonstrates | Command |
| --- | --- | --- |
| [Text response](examples/respond/main.go) | A prompt with instructions and interrupt handling | `go run ./examples/respond "What does chmod do?"` |
| [Structured output](examples/schema/main.go) | Generate a schema, request JSON, and decode into a Go struct | `go run ./examples/schema` |
| [Image prompt](examples/image/main.go) | Attach an image to a prompt | `go run ./examples/image Screenshot.png` |

These are runnable examples, not recorded native verification. The schema example also exercises provisional `--text` and integer-field flags; see [compatibility](docs/compatibility.md).

## API overview

| Method | Result |
| --- | --- |
| `client.Available(ctx)` | Readiness and raw availability output |
| `client.Respond(ctx, prompt, opts)` | Buffered stdout, stderr, and optional JSON |
| `client.CountTokens(ctx, request)` | Parsed token count |
| `client.SchemaObject(ctx, schema)` | Schema JSON to write to a file and use with `SchemaPath` |

Browse [package source](pkg/fm) or run `go doc ./pkg/fm` in a checkout for the exported API.

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

## Compatibility

Reviewed against Apple's published documentation on **2026-10-29**. This SDK wraps CLI operations; it does not expose the complete Foundation Models framework.

Apple's [Foundation Models updates](https://developer.apple.com/documentation/updates/foundationmodels) describe image analysis, Dynamic Profiles, tool-calling controls, model APIs, and revised Swift error types. Those APIs do not establish equivalent CLI flags or exit codes. Availability, token counting, some flag forms, and error categories still await native verification.

See [the compatibility review](docs/compatibility.md) for the source audit, unsupported capabilities, and commands to verify before a release.

## Provisional CLI mapping

| Go | `fm` arguments |
| --- | --- |
| `Available` | `available` |
| `Respond` | `respond [--instructions=…] [--model=…] [--image=…]… [--schema=…] (--text=… \| -- <prompt>)` |
| `CountTokens` | `count-tokens --text=…` |
| `SchemaObject` | `schema object --name=… --string=… --int=… [--array]` |

Flag names live in `pkg/fm/flags.go`; argument ordering and output parsing live in the corresponding operation files. The [compatibility review](docs/compatibility.md) separates Apple-documented spellings from assumptions. In particular, `available`, `count-tokens`, `--text`, `--int`, and error prefixes still require native confirmation.

## Development

Use [Just](https://github.com/casey/just) for the repository's checks:

```sh
just --list                # Discover available commands
just test                  # Isolated fake-process tests with the race detector
just lint                  # Formatting, vet for both tags, optional golangci-lint
just check cross-build     # Linux/macOS/Windows × amd64/arm64, without CGo
just test-applefm           # Native smoke tests; unavailable prerequisites skip
just check native-required # Native release gate; missing prerequisites fail
```

The unit suite runs in a disposable Docker container with read-only source and networking disabled. Docker and the `golang:1.24-bookworm` image are required. Checks live in the modular [.justfiles/check.just](.justfiles/check.just); [CI](https://github.com/Obedience-Corp/apple-fm-go/actions/workflows/ci.yml) uses the same recipes.

Native tests check availability, a nonempty on-device response, schema JSON, and token counts with bounded contexts. A skipped run does not establish compatibility. Image/PCC behavior and schema round-trips require separate native verification.

## Contributing

For CLI compatibility fixes, include the macOS version, relevant `fm --help` output, and a minimal reproduction with private inputs removed. Update the [compatibility review](docs/compatibility.md), add a regression test for changed behavior, and run the checks above. Keep new APIs grounded in observed CLI behavior.

## Related

- [obey-go-sdk](https://github.com/Obedience-Corp/obey-go-sdk) — Go client for the Obey daemon.
- [claude-code-go](https://github.com/lancekrogers/claude-code-go) — Claude Code CLI SDK.
- [muse-spark-go](https://github.com/Obedience-Corp/muse-spark-go) — Muse CLI and Session Protocol SDK.

## License

Apache-2.0. See [LICENSE](LICENSE).
