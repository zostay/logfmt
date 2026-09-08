---
kind: cli
forge: github
tracker: github-issues
test: make test
deploy.mode: manual-tag
deploy.target: github-releases
---

# logfmt

A Go command-line tool that reformats and colorizes mixed text/JSON log streams
(`kubectl logs deploy/app -f | logfmt`). Single `main` package, installed with
`go install ./`.

Merging to `master` publishes nothing. Releases are cut deliberately: a
`release/vX.Y.Z` branch dry-runs the version, changelog and cross-builds via
`.github/workflows/prepare.yaml`, and only pushing the `vX.Y.Z` tag afterwards
runs `.github/workflows/release.yaml`, which builds the three binaries
(linux/amd64, darwin/arm64, darwin/amd64) and publishes the GitHub release.
The `/release` skill in `.claude/skills/release` drives that, and it asks a
human to confirm the version number, so cutting a release is not unattended
work.

Two release checks are strict enough to be worth knowing in advance:
`version.txt` must contain the version, and the first line of `Changes.md` must
be exactly `## X.Y.Z  YYYY-MM-DD` (two spaces) with the date being *today in
America/Chicago* at the moment the workflow runs — so a release straddling the
Central midnight fails at the tag even though the branch passed.

`make test` runs `go test ./...`. CI also runs `golangci-lint` (pinned to
v2.4.0) and a coverage gate that is currently set to `REQUIRED_COVERAGE: 0`, so
coverage is reported but not enforced.
