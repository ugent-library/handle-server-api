# Plan: rebuild the container image with current Go and CA certificates

**Status:** in progress. Copied from resolver's archived plan (`resolver/docs/archive/docker-image-build.md`) on 2026-10-29. Ingest and resolver have both done part A; their answers to the open questions are noted below as precedent.

## Reusing this plan in another project

This plan is written to be copied to other Go projects that build an Alpine image on Docker Hub and run on Nomad. Everything outside *This project* and *Part B* applies to such a project as it stands. To reuse it:

1. Copy this file to the other project's `docs/plans/`.
2. Fill in *This project* with that project's values. The steps in part A refer to them by name (the binary path, the build and test commands).
3. Replace *Part B* with that project's own cleanup, or delete it.
4. Work through part A there and record the answers to the open questions in that copy.

There is no need to repeat the test in *What we tested*: it shows how Go's `crypto/x509` and Alpine's CA bundle behave, which is the same for every Go binary on Alpine.

## This project

| | handle-server-api |
| --- | --- |
| Outbound HTTPS calls | none. The only outbound connection is to Postgres through pgx (`HDL_DSN`). The default DSN uses `sslmode=disable`; production's sets no `sslmode` (see *Exposure*) |
| Build stage today | `FROM golang:1.24.0-alpine AS build`, runs `go get -d -v ./...` then `go build -buildvcs=false -o handle-server-api -v` |
| Final stage today | `FROM alpine:latest`, no package upgrade; copies only the binary to `/dist` |
| Binary in the image | `/dist/handle-server-api` |
| `go` directive in go.mod | `go 1.24.0` |
| `hooks/build` | does not exist |
| Build and test commands | `go build -buildvcs=false -o handle-server-api`, `go vet ./...` (there are no tests and no Makefile) |
| Running containers | *to look up* (number of Nomad allocations) |
| Branches | `master` is the main branch (GitHub default). `main` and `dev` also exist, all three at `13a47e6` on 2026-10-29; also `pg` and a dependabot branch for pgx 5.9.2. Moving to `main` only is a follow-up issue |
| Merge commit message | `merge <branch>: <summary>`, lowercase like the rest of the history (the convention set in resolver) |
| Releases | git tags `v0.0.1` to `v0.0.6`, no GitHub releases checked yet; `v0.0.6` is on `13a47e6` |
| Docker Hub build rules | branch `dev` → tag `dev`, branch `master` → tag `master`, git tag `vX.Y.Z` → tag `release-vX.Y.Z` (looked up 2026-10-29). Which image tag production runs is *to confirm*; if it is a `release-*` tag, A6 needs a new git tag |

The Dockerfile declares `ARG SOURCE_BRANCH`, `SOURCE_COMMIT` and `IMAGE_NAME` and copies them into `ENV`, but without a `hooks/build` Docker Hub does not pass them, so they are empty in the running image. The `/info` endpoint returns them (`config.Version`), so today it answers `{}`. The hook added in A3 fills them in.

**Exposure.** handle-server-api makes no outbound HTTPS calls. The one TLS connection it can make is to Postgres: pgx verifies the server certificate against the image's CA bundle with `sslmode=verify-ca` or `verify-full`. With `disable`, `allow`, `prefer` or `require` it does not verify the chain, so the eMudhra compromise in *Why now* does not affect it. *Production's `HDL_DSN` sets no `sslmode` (2026-10-29), so pgx uses its default, `prefer`: it tries TLS without verifying the certificate and falls back to plain. No chain is checked, so the service is not exposed.* Either way, Go 1.24 no longer gets security fixes, so moving to a supported Go release is reason enough to do this plan.

## Why now

On 17 October 2026 a researcher reported that the private keys of eight eMudhra intermediate CA certificates could be recovered from their public certificates. Anyone holding those keys can issue a TLS certificate that a client trusts for any domain. Browsers blocked the intermediates between 17 and 20 October, and the root programs distrusted `emSign ECC Root CA - G3` on 21 October. A Go service is not a browser: every outbound HTTPS call trusts whatever Go's `crypto/x509` and the image's CA bundle trust. Go does not check revocation, so the revocation of the intermediates on 18 and 19 October does not protect us.

The [disclosure site](https://emudhra.we-have-a-problem.com/) publishes the affected certificates in `impacted-intermediates.json`: sixteen certificates for the eight keys. Each intermediate exists twice, once issued by an emSign root and once cross-signed by Sectigo's `AAA Certificate Services`, and both copies carry the same key.

Two upstream fixes exist, and each one closes the hole in an Alpine image on its own:

- **Go 1.27.3** (released 2026-10-19, [CL 851047](https://go-review.googlesource.com/c/go/+/851047), CVE-2026-111943) makes `x509.Certificate.Verify` drop chains through the compromised intermediates, identified by the SHA-256 hash of their public key. That covers both copies of every intermediate, whatever root the chain ends in. The fix is compiled into the binary. Go 1.26.10 carries the same fix for projects still on Go 1.26.
- **Alpine `ca-certificates-bundle 20261020-r0`** (v3.24, v3.23, v3.22 and edge) removes `emSign ECC Root CA - G3`, following Mozilla's NSS 3.131.1.

### What we tested (2026-10-26)

A small Go HTTPS client, built once with Go 1.27.2 and once with Go 1.27.3, run in Alpine images with the old (`20260909-r0`) and the new (`20261020-r0`) bundle, against the four test hosts. The bundle contents were checked by SHA-256 fingerprint:

| Bundle | `emSign ECC Root CA - G3` | `emSign Root TLS CA - G3` | `AAA Certificate Services` |
| --- | --- | --- | --- |
| `20260909-r0` | present | absent | absent |
| `20261020-r0` | absent | absent | absent |

| Go | Bundle | emSign chain | Sectigo chain |
| --- | --- | --- | --- |
| 1.27.2 | old | **accepted** | rejected |
| 1.27.3 | old | rejected (`no valid chains built`) | rejected |
| 1.27.2 | new | rejected (`certificate signed by unknown authority`) | rejected |
| 1.27.3 | new | rejected | rejected |

The hosts that serve certificates never logged to Certificate Transparency gave the same results, since Go does not enforce CT. So:

- An image is exposed only with an old Go and an old bundle together. Any image built before 20 October most likely has both.
- The Sectigo cross-signature does not reach an Alpine image: Alpine's bundle, built from Mozilla's, does not contain `AAA Certificate Services`. It matters for trust stores that do contain it, such as those of macOS and Windows, which is one reason the Go fix is worth having alongside the bundle.
- The error text shows which fix rejected a chain: `no valid chains built` is Go's filter, `certificate signed by unknown authority` is the bundle.

We take both fixes, so that a later mistake in one (a stale bundle, a toolchain left behind) does not reopen the hole. Because each fix was shown to work, the checks below only confirm the versions: Go 1.27.3 in the binary and bundle `20261020-r0` or later in the image.

## Why a rebuild may pick up neither

- `FROM golang:alpine` and `FROM alpine:latest` are floating tags. Docker uses whatever copy of the tag it already has unless the build runs with `--pull`, so a rebuild can produce a binary built with an older Go, on an old bundle. *handle-server-api pins the build stage (`golang:1.24.0-alpine`), so a rebuild keeps Go 1.24.0 until the tag is changed; the final stage floats.*
- Without a package upgrade in the final stage, the image has whatever bundle its base image shipped with. And a `RUN apk upgrade` alone does not help for long: Docker reuses a `RUN` layer as long as its parent layer and its command text are unchanged; it never looks at the Alpine package index. The layer would run once and stay cached.
- Docker Hub runs `hooks/build` instead of its own `docker build` when that file exists. Its default build does not pass `--build-arg`, so any `ARG` in the Dockerfile stays empty.

**Model:** `ugent-library/depot-management` solves the caching problem. Its `hooks/build` passes `--build-arg SNAPSHOT_EPOCH="$(date +%s)"`, and its Dockerfile declares `ARG SNAPSHOT_EPOCH` right before the package-install `RUN`. A declared `ARG` is part of the environment of every `RUN` after it in the same stage, so a new value on every build makes that `RUN` miss the cache and run again.

## Open questions

- **Periodic rebuilds.** Docker Hub builds only on a push or a manual trigger. `SNAPSHOT_EPOCH` makes a build that runs pick up new packages; it does not make a build happen. If a project goes months without a push, its image keeps an old Go and an old bundle. Decide how rebuilds happen between code changes: by hand after each Go point release and Alpine security update, or on a schedule from outside Docker Hub. Check what Docker Hub currently offers for triggering builds before choosing. *Ingest chose by hand; resolver moved the question to its `docs/plans/docker-image-slim.md`. handle-server-api: open.*
- **Raise the `go` directive in go.mod?** Raising it to `go 1.27.3` makes a build on an older toolchain fail outright, because the official golang images set `GOTOOLCHAIN=local`. It also forces every developer onto 1.27.3 (Go downloads it automatically unless `GOTOOLCHAIN=local` is set locally). Pinning the image tag already fixes the Docker build; this would add a check that stops a stale toolchain anywhere. *Ingest and resolver both raised it; resolver also ran `go get -u ./...` and `go mod tidy` and compared the old and new build side by side before merging. Of the GODEBUG defaults the new directive switches on, only `containermaxprocs`/`updatemaxprocs` mattered there: `GOMAXPROCS` follows a hard CPU limit on the container, if the Nomad job sets one. handle-server-api: raised to `go 1.27.3` (decided 2026-10-29), without updating dependencies; the pgx update is a follow-up issue. Test that the build compiles and the API behaves as before.*
- **Docker Hub "Build caching" setting.** Look up whether it is on for the repository (Builds, Configure automated builds). If it is off, every build already starts from scratch and the EPOCH argument only matters if caching is turned on later. Keep the argument either way, so the Dockerfile does not depend on that setting. *Resolver: on. handle-server-api: to look up.*

## Part A: pick up both fixes

The fix branches from `master`, not `dev`, so that releasing it does not also release whatever is waiting on `dev`. It is merged into both, so `dev` does not lose it at the next release. *On 2026-10-29 `master`, `main` and `dev` are the same commit, so nothing is waiting on `dev` yet. `main` is left alone; retiring `master` in favour of `main` is a follow-up issue.*

### A0. Branch

- [x] Switch to `master` and bring it up to date: `git switch master && git pull --ff-only`.
- [x] Create the feature branch: `git switch -c docker-image-build`. Steps A1 to A3 are commits on this branch.

### A1. Go

- [x] Change the build stage to `FROM golang:1.27.3-alpine AS build`.
- [x] Decide the go.mod question above; if yes, change the `go` directive to `1.27.3` in the same commit. *handle-server-api: raised with `go mod edit -go=1.27.3` and `go mod tidy`. Tidy changed no versions: it moved `alexliesenfeld/health` from indirect to direct (it has been imported directly since the health check was added) and added the `stretchr/objx` checksums to go.sum.*
- [x] Update *Build requirements* in the README (it says Golang v1.24.0).
- [ ] Build locally with `docker build --pull` and check which toolchain built the binary: copy it out with `docker create` and `docker cp <container>:/dist/handle-server-api ./handle-server-api`, then run `go version ./handle-server-api` on the host. It must print `go1.27.3`. *handle-server-api (2026-10-29): `go1.27.3`; the image built from `master` prints `go1.24.0`.*

### A2. CA bundle, kept current

- [x] Change the final stage to `FROM alpine:3.24`, the release `alpine:latest` points to today. Pinning it makes moving to a new Alpine release a deliberate change rather than something a rebuild does unnoticed.
- [x] Right after that `FROM`, add the argument and the upgrade. The `ARG` must sit inside the final stage; one declared before the first `FROM` is not visible there. Upgrade all packages rather than naming one: the base image ships `ca-certificates-bundle`, not `ca-certificates`, so `apk upgrade ca-certificates` would change nothing, and the other packages get their security fixes too.

  ```dockerfile
  FROM alpine:3.24
  # hooks/build passes a new value on every build, so this layer is never served from cache.
  ARG SNAPSHOT_EPOCH
  RUN apk upgrade --no-cache
  ```

- [ ] Check the result: `docker run --rm --entrypoint apk <image> list -I ca-certificates-bundle` shows `ca-certificates-bundle-20261020-r0` or later. Grepping the bundle for "emSign" proves nothing: `/etc/ssl/certs/ca-certificates.crt` holds PEM blocks only, without the root names, so the count is always 0. *handle-server-api (2026-10-29): `ca-certificates-bundle-20261020-r0` on Alpine 3.24.3. A local rebuild of `master` shows the same bundle, but only because `alpine:latest` was pulled fresh; the image production runs now may be older.*

### A3. Docker Hub build hook

- [x] Add `hooks/build`, modelled on depot-management's. It fills in `SOURCE_BRANCH` and `SOURCE_COMMIT` from git when Docker Hub leaves them empty. Leave out the `--build-arg` lines for arguments the Dockerfile does not declare.

  ```bash
  #!/bin/bash
  if [[ -z "$SOURCE_BRANCH" ]]; then
  	export SOURCE_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
  fi
  if [[ -z "$SOURCE_COMMIT" ]]; then
  	export SOURCE_COMMIT="$(git rev-parse HEAD)"
  fi

  echo "SOURCE_BRANCH: $SOURCE_BRANCH"
  echo "SOURCE_COMMIT: $SOURCE_COMMIT"
  echo "IMAGE_NAME: $IMAGE_NAME"

  docker build \
      --pull \
      --build-arg SOURCE_BRANCH=$SOURCE_BRANCH \
      --build-arg SOURCE_COMMIT=$SOURCE_COMMIT \
      --build-arg IMAGE_NAME=$IMAGE_NAME \
      --build-arg SNAPSHOT_EPOCH="$(date +%s)" \
      -f $DOCKERFILE_PATH \
      -t $IMAGE_NAME .
  ```

- [x] Make it executable (`chmod +x hooks/build`) and commit the mode.
- [x] Run the hook locally with `DOCKERFILE_PATH=Dockerfile IMAGE_NAME=handle-server-api:local hooks/build` to check it works outside Docker Hub: it prints the three values, pulls both base images, and runs `apk upgrade` rather than showing `CACHED`. *handle-server-api (2026-10-29): on a second run the `go build` layer was `CACHED` and `apk upgrade` ran again; `env` in the image shows `SOURCE_BRANCH`, `SOURCE_COMMIT` and `IMAGE_NAME`.*
- [ ] Look up the Docker Hub "Build caching" setting and record the answer in this plan.

### A4. Check locally

- [x] Run the build and test commands from *This project* with the Go 1.27.3 toolchain. Also run `go mod verify` and `gofmt -l .`. *handle-server-api (2026-10-29): `go build` and `go vet ./...` pass, all modules verified, `gofmt` clean.*
- [x] Run the image against a local Postgres with the handle schema (`HDL_DSN` and `HDL_PREFIX` as in `.env.example`): `/status` answers, and a `PUT`, `GET` and `DELETE` on `/handles/{prefix}/{local_id}` with basic auth behave as the README describes. Check that `env` in the container shows `SOURCE_BRANCH` and `SOURCE_COMMIT`. *handle-server-api (2026-10-29): `postgres:16-alpine` with the handle.net `handles` table (bytea columns, primary key `handle, idx`), one seeded handle with a URL and an `HS_ADMIN` row, DSN without `sslmode` as in production. The image built from `master` and the new image each got the same 17 requests: `/status`, `/info`, no auth, wrong password, wrong prefix, `GET` missing, `PUT` without `url`, `PUT` new, `GET`, `PUT` replace, `GET`, `GET` seeded, `DELETE`, `GET` deleted, `DELETE` again, `POST` (405), unknown path. The responses are identical (status, content type, body with timestamps masked) except `/info`, which answered `{}` before and now shows branch, commit and image. No errors or panics in either log; the seeded rows were untouched.*
- [ ] Look up the Docker Hub build rules, the running containers and production's DSN mode, and record them in *This project*. Steps A5 and A6 depend on them. If no rule builds `dev`, or there is no development environment, A4 is the only test before production; then also test against a copy of the production database, or with production's `sslmode` if it verifies certificates.

### A5. Merge into dev and test there

- [ ] Merge the feature branch into `dev` with a merge commit: `git switch dev && git pull --ff-only && git merge --no-ff docker-image-build`, with the merge commit message from *This project*.
- [ ] Push `dev`. Docker Hub builds from what is pushed, so the development environment can only be tested after this push.
- [ ] In the Docker Hub build log, check that the hook ran (it prints `SOURCE_BRANCH`, `SOURCE_COMMIT` and `IMAGE_NAME`) and that the `apk upgrade` step ran rather than showing `CACHED`.
- [ ] Pull the image Docker Hub built and repeat the checks from A1 and A2: `go version` on the binary and the bundle version. Also check that `env` shows non-empty `SOURCE_BRANCH` and `SOURCE_COMMIT`.
- [ ] Deploy to the development environment. Open a shell in every running allocation and run the command below. Type the path itself: a placeholder such as `<binary>` makes the shell read `<` as a redirect and fail with a syntax error.

  ```sh
  grep -a -o -E -m1 "go1\.[0-9]+\.[0-9]+" /dist/handle-server-api | head -1; apk list -I ca-certificates-bundle
  ```

  It must print `go1.27.3` and `ca-certificates-bundle-20261020-r0` or later. The image has no Go toolchain, but the binary holds its Go version as text, which BusyBox `grep` can read.
- [ ] Check that the service works as before on the development environment: `/status`, and a `GET` on a known handle.
- [ ] Trigger a second Docker Hub build of `dev` without code changes and check that the `apk upgrade` step runs again. Skip this if Build caching is off.

### A6. Merge into master and release

- [ ] Merge the feature branch into `master` with a merge commit: `git switch master && git pull --ff-only && git merge --no-ff docker-image-build`, with the merge commit message from *This project*. Merge the feature branch, not `dev`, so nothing else from `dev` comes along.
- [ ] Push `master`.
- [ ] Decide whether a release is needed. It is if production runs an image built from a tag: a new tag is then the only way to get the new Dockerfile into production, because rebuilding the old tag rebuilds the old Dockerfile. If production runs the image built from `master`, the push is enough.
- [ ] If a release is needed: create `v0.0.7` as an annotated tag on the merge commit on `master`, push the tag, and write a GitHub release that says the image is rebuilt with Go 1.27.3 and the updated CA bundle to move off unsupported Go 1.24 and pick up current Alpine packages, with no other changes and no upgrade steps.
- [ ] Deploy to production and run the same check in every running allocation as in A5.
- [ ] Delete the feature branch, locally and on GitHub if it was pushed.
- [ ] Decide the periodic rebuild question and record the answer here.

## Part B

None. The final stage already copies only the binary, so the cleanup resolver needed does not apply here.

## Follow-up GitHub issues

To create once this plan is done:

- [ ] Move to `main` as the only branch and deprecate `master` (Docker Hub build rules, GitHub default branch, the `dev` merge flow).
- [ ] Update the Go dependencies, starting with pgx (dependabot branch for pgx 5.9.2).

## Retiring the plan

- [ ] Describe how the image is built (Docker Hub, `hooks/build`, the `SNAPSHOT_EPOCH` argument, pinned base images, how rebuilds are triggered) in the README, so the next person to bump Go or Alpine knows which lines to change.
- [ ] Move this plan to `docs/archive/` or delete it.
