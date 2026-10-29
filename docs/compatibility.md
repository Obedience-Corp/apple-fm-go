# Foundation Models compatibility review

Reviewed 2026-10-29 against the wrapper merged in PR #1 (`a2b0ab2`).

## Evidence

| Source | What it establishes | Consequence for this module |
| --- | --- | --- |
| [Apple's CLI session, code at 5:07 and 7:55](https://developer.apple.com/videos/play/wwdc2026/334/) | `respond`, `--instructions`, `--model pcc`, `--image`, `--schema`, and `schema object --name … --string … --array` | Existing argument builders cover these operations. Remove the README's incorrect PCC exclusion. |
| [Foundation Models updates](https://developer.apple.com/documentation/updates/foundationmodels) | Latest listed update is June 2026; model and session APIs have evolved | Do not infer CLI flags, context limits, or exit codes from Swift API names. |
| [Apple Python SDK](https://github.com/apple/python-apple-fm-sdk) | A separate framework binding with a macOS 26 deployment floor | This Go subprocess wrapper still requires the macOS 27 executable. |
| [Python SDK change #20](https://github.com/apple/python-apple-fm-sdk/commit/0249f35fce855a6701debaed569b6827fc6f5706) | October 7 adds model variant display names | No equivalent CLI discovery API is established by this change. |

The checked update page and Python repository do not show a newer change since this wrapper's October 21 scaffold. This is a review of those sources, not proof that every installed OS build has identical behavior.

## Remaining CLI assumptions

No native help or output fixture has been captured for this repository. In particular, these existing behaviors remain provisional:

- `available`, its exit status, and the license setup procedure.
- `count-tokens --text=…` and its single-integer output.
- `respond --text=…`, the explicit model name `system`, and integer schema fields.
- `--flag=value`, repeated image flags, and `--` before positional prompts.
- The stderr prefix categories in `flags.go`; no verified exit-code mapping exists.

The examples Apple publishes use separate flag/value arguments. The wrapper's equals form is tested against a fake subprocess only. Keep option-like input protected until real help/output proves the parser contract; do not change argument handling based on a guess.

The wrapper buffers stdout/stderr and starts a fresh process per call. It does not implement streaming callbacks, chat history, tool execution, Swift Dynamic Profiles, remote-provider configuration, or a Python bridge. New framework error types cannot safely be mapped to CLI stderr without evidence.

## Native validation before v0.1.0

On an eligible macOS 27 Mac, record `sw_vers` and the outputs of:

```sh
fm --help
fm respond --help
fm schema object --help
fm available --help
fm count-tokens --help
just check native-required
```

Follow the installed binary's license/setup instructions manually. Never accept terms from library code or automate an OS upgrade as part of these checks.

The native tests have per-operation timeouts and check availability, nonempty text, valid schema JSON, and positive token counts. They deliberately fail on unknown CLI errors. The strict recipe fails instead of skipping missing/unavailable prerequisites. They do not write schema files or run the fake-process filesystem fixtures on the host.

Then verify a schema file round-trip with `examples/schema`, an image response with `examples/image`, and explicit PCC selection if intended for the release. Record model/OS details and errors without private inputs. Correct argument builders and parsing where necessary, and retain sanitized fixtures. PCC access is a separate check; passing the on-device test does not establish it.

Run `just test`, `just lint`, and `just check cross-build` after corrections. The fake suite covers cancellation, Unix process-group cleanup, missing binaries, unclassified exits, invalid JSON, argument validation, and environment propagation; none of those tests establishes Apple's runtime compatibility.
