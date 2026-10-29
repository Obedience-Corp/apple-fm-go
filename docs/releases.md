# Releasing apple-fm-go

This repository publishes a Go module. Releases contain source archives supplied by GitHub; there are no CLI binaries or platform-specific SDK assets to upload.

## Beta releases

Tags use `vMAJOR.MINOR.PATCH-beta.N`, starting at `v0.1.0-beta.1`. The [release workflow](../.github/workflows/release.yml) runs only for beta tags. It checks tag syntax, main-branch ancestry, release notes, changelog, lint, isolated race tests, module integrity, and all six build targets before publishing a GitHub **prerelease**. It never marks a beta as Latest. Only the publication job receives write permission.

1. Add `docs/releases/<tag>.md` with features, validation evidence, and limitations. Add an exact `## [<version-without-v>]` heading to `CHANGELOG.md`.
2. Merge the preparation PR into `main`, including the workflow before the first automated release.
3. Fetch and check out the merged commit in a clean project worktree. Run `just release check`. Review the diff and notes for that exact commit.
4. Create an annotated tag and push only that tag:

   ```sh
   git tag -a v0.1.0-beta.2 -m 'Release v0.1.0-beta.2'
   git push origin refs/tags/v0.1.0-beta.2
   ```

5. Verify the workflow succeeded and the GitHub release is marked prerelease. Check that a clean consumer can resolve the exact module version.

Never move a published tag. Fix failures on a new commit and use the next beta version when the source needs to change. If only the workflow infrastructure failed, rerun the failed jobs after resolving the cause.

`v0.1.0-beta.1` bootstrapped the process: it was published manually from already-merged `main`, before the release workflow landed. The tag's README consequently predates the public installation instructions in this PR.

## Stable release gate

The beta workflow deliberately does not publish stable tags. Before adding stable automation or tagging `v0.1.0`, capture real macOS 27 CLI help and verify `just check native-required`, schema round-trips, and the claimed image/model behavior. Record sanitized evidence and correct the SDK against the installed binary. See [compatibility](compatibility.md).

Passing fake-process tests, successful cross-builds, or a skipped native test do not satisfy that gate.
