---
name: release
description: Cut a release — create a release/vX.Y.Z branch that dry-runs the release build via the Prepare workflow, merge it once checks pass, then tag vX.Y.Z to trigger the real Release workflow. Invoke as `/release [version]`.
---

# Release

Cut a release of logfmt. The release is driven by two GitHub workflows that
validate the same things, so a mistake is caught on the release branch (by
`prepare.yaml`) before it can reach the tag (`release.yaml`).

- `.github/workflows/prepare.yaml` — runs on push to `release/*`. Checks the
  version and changelog, then cross-builds all three binaries. This is the dry
  run.
- `.github/workflows/release.yaml` — runs on push of a `v*` tag. Repeats the
  same checks and builds, then creates the GitHub release, uploads the binaries,
  and un-drafts it.

## The three things the workflows enforce

Get these wrong and the workflow fails. They are the whole reason the release
branch exists.

1. **`version.txt` must contain the release version.** The check is
   `grep -q "$RELEASE_VERSION" version.txt`.
2. **The first line of `Changes.md` must be exactly `## X.Y.Z  YYYY-MM-DD`** —
   two spaces between version and date. Anything else, including a leftover
   `## Unreleased` section above it, fails the check.
3. **That date must be today in `America/Chicago`.** Both workflows set the
   timezone to US Central and compare against `date "+%Y-%m-%d"` *at the moment
   they run*. This is the one that bites: if the Central date rolls over between
   merging the branch and pushing the tag, the Release workflow fails even
   though Prepare passed. See "If the date rolls over" below.

The release notes are everything between the first `## <digit>` heading and the
next one, so the consolidated section for this version becomes the release body.

## Steps

### 1. Preflight

```bash
git status --porcelain          # must be clean
git rev-parse --abbrev-ref HEAD # must be master
git pull --ff-only
gh run list --branch master --limit 1   # master should be green
go test ./... && golangci-lint run
```

Stop and tell the user if the tree is dirty or master is red. Do not release
from a branch other than `master`. If the master run is still `in_progress`,
either wait for it or note explicitly that you proceeded without it.

Confirm there is something to release: `Changes.md` should have at least one
`## Unreleased` section. If it does not, stop — there is nothing to cut.

### 2. Choose the version

Read the current version:

```bash
cat version.txt
git tag --sort=-v:refname | head -3
```

If the user passed a version as an argument, use it. Otherwise propose the next
one from the unreleased entries — behavior changes or new features mean a minor
bump on this 0.x project, pure fixes mean a patch bump — and **ask the user to
confirm before proceeding**. The version becomes a permanent public tag and
release, so do not guess silently.

Use the bare `X.Y.Z` form for `version.txt` and the changelog heading, and the
`vX.Y.Z` form for the branch name and the tag.

### 3. Create the release branch

```bash
git checkout -b release/vX.Y.Z
```

The branch name must contain the version — both workflows extract it from the
ref with `grep -Eo '[0-9]+\.[0-9]+\.[0-9]+.*$'`.

### 4. Update the version and changelog

Write the bare version into `version.txt` (keep the trailing newline). Use `>|`
rather than `>`: this shell has `noclobber` set, so a plain `>` onto an existing
file stops and waits for an interactive `overwrite?` answer.

```bash
printf '%s\n' "X.Y.Z" >| version.txt
```

Then **consolidate every `## Unreleased` section in `Changes.md` into a single
`## X.Y.Z  <today>` section at the top.** Get the date from Central time, not
local time:

```bash
TZ=America/Chicago date +%Y-%m-%d
```

Merge the bullets from all the Unreleased sections in chronological order, and
edit them into release notes a user would want to read: drop noise, group
related entries, and keep the wording user-facing. These bullets are published
verbatim as the GitHub release body. No `## Unreleased` heading may remain above
the new version heading.

Verify before committing, by running the same checks the workflows run:

```bash
RELEASE_VERSION=X.Y.Z
grep -q "$RELEASE_VERSION" version.txt && echo "version.txt PASS"
date=$(TZ=America/Chicago date "+%Y-%m-%d")
[ "$(head -n1 Changes.md)" = "## $RELEASE_VERSION  $date" ] && echo "heading PASS"
grep -c '^## Unreleased' Changes.md    # must be 0
```

Preview the release body — this is published verbatim, so read it before it
ships. The workflow's `sed` script is GNU-only and fails on macOS, so use awk
locally:

```bash
awk '/^## [0-9]/{n++; if(n==2) exit; next} n==1' Changes.md
```

Then dry-run the three cross-builds locally, so a build failure surfaces here
instead of in CI. Note that zsh does not word-split unquoted parameters, so
write these out rather than looping over `"linux amd64"`-style pairs:

```bash
GOOS=linux  GOARCH=amd64 go build -o /tmp/logfmt-linux-amd64  ./
GOOS=darwin GOARCH=arm64 go build -o /tmp/logfmt-darwin-arm64 ./
GOOS=darwin GOARCH=amd64 go build -o /tmp/logfmt-darwin-amd64 ./
go run . --version    # should print the new version
```

Remove any stray `logfmt` binary a bare `go build ./...` leaves in the repo root
before committing — it is not gitignored.

### 5. Commit and push the release branch

```bash
git add version.txt Changes.md
git commit -m "chore(releng): version and changes"
git push -u origin release/vX.Y.Z
```

Use that exact commit message — it is the convention for every prior release.

The push triggers `prepare.yaml`. Also open a PR so the test suite runs too
(`test.yaml` runs on `pull_request`, not on `release/*` pushes):

```bash
gh pr create --base master --title "Release vX.Y.Z" --body "<summary of the release>"
```

### 6. Wait for the checks

Both must pass before merging:

```bash
gh pr checks <number> --watch
gh run list --branch release/vX.Y.Z --limit 5
```

- **Prepare for Release** — the dry run. If it fails, fix the cause on the
  branch and push again; the whole point is to fail here rather than at the tag.
- **Test and Sanity** — the normal suite.

### 7. Merge

```bash
gh pr merge <number> --merge --delete-branch
git checkout master && git pull --ff-only
```

Use a merge commit, matching every prior release.

### 8. Tag the release

Tag the merge commit on `master` and push the tag:

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

This triggers `release.yaml`, which builds the binaries, creates the release as
a draft, uploads the three artifacts, and then un-drafts it.

### 9. Verify the release

```bash
gh run list --workflow=Release --limit 1
gh run watch <run-id>
gh release view vX.Y.Z
```

Confirm the release exists, **is not a draft**, and has all three binaries:

- `logfmt-X.Y.Z-linux-amd64`
- `logfmt-X.Y.Z-darwin-arm64`
- `logfmt-X.Y.Z-darwin-amd64`

Then confirm the published artifact actually works, rather than trusting that a
green workflow means a good binary:

```bash
gh release download vX.Y.Z -p 'logfmt-X.Y.Z-darwin-arm64' -D /tmp/rel
chmod +x /tmp/rel/logfmt-X.Y.Z-darwin-arm64
/tmp/rel/logfmt-X.Y.Z-darwin-arm64 --version    # must print the new version
```

Check the release body matches the changelog section:

```bash
gh release view vX.Y.Z --json body --jq .body | sed '/^$/d' > /tmp/body.txt
awk '/^## [0-9]/{n++; if(n==2) exit; next} n==1' Changes.md | sed '/^$/d' | diff - /tmp/body.txt
```

### 10. Report

Report the version, the PR, the tag, the release URL, and the artifacts. Note
anything that needed a retry.

## If the date rolls over

If the Central date changes between the Prepare run and the tag push, the
Release workflow fails its changelog date check. Fix it on `master`:

1. Update the date in the `Changes.md` heading to the new Central date.
2. Commit to `master` (or via a PR, matching repo convention).
3. Delete and re-push the tag so it points at the corrected commit:

```bash
git tag -d vX.Y.Z
git push origin :refs/tags/vX.Y.Z
git tag vX.Y.Z
git push origin vX.Y.Z
```

If a draft release was already created by the failed run, delete it first with
`gh release delete vX.Y.Z` so the retry can recreate it cleanly.

## Notes

- Never push a `v*` tag without going through the release branch first. The tag
  is what publishes to the public, and the branch is the only dry run.
- Do not amend or force-push a tag that has already produced a published
  release; cut a new patch version instead.
