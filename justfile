# host-first: native-macos
set positional-arguments
set shell := ["bash", "-euo", "pipefail", "-c"]

export REPO_DIR := justfile_directory()
export COMPOSE_PROJECT_NAME := "repo-" + sha256(justfile_directory())
export GIT_COMMON_DIR := `git rev-parse --path-format=absolute --git-common-dir`

# Show available recipes
_default:
    @just --list

# Build dev container image
build:
    @docker volume create rubio-uv-cache >/dev/null
    @docker volume create rubio-npm-cache >/dev/null
    @docker volume create rubio-gomod-cache >/dev/null
    @docker volume create rubio-cargo-registry >/dev/null
    @docker volume create rubio-coursier-cache >/dev/null
    docker compose build --pull dev

# Build the dev image, sync dependencies into volumes, and install git hooks
setup: build
    bin/setup.sh
    @cid="$(docker create ghcr.io/rubio-enterprises/base:main)"; trap 'docker rm "$cid" >/dev/null' EXIT; mkdir -p .standards; docker cp "$cid:/opt/standards/." .standards/
    @HOOKS_DIR="$(git rev-parse --git-path hooks)"; \
    mkdir -p "$HOOKS_DIR"; \
    rm -f "$HOOKS_DIR/post-checkout" "$HOOKS_DIR/post-merge" "$HOOKS_DIR/post-rewrite"; \
    for hook in pre-commit commit-msg pre-push; do \
        printf '#!/bin/sh\nset -eu\nREPO_DIR="$(git rev-parse --show-toplevel)"\nexport REPO_DIR\nexport GIT_COMMON_DIR="$(git -C "$REPO_DIR" rev-parse --path-format=absolute --git-common-dir)"\nexport COMPOSE_PROJECT_NAME="repo-$(printf %%s "$REPO_DIR" | shasum -a 256 | cut -d " " -f 1)"\ncd "$REPO_DIR" && exec docker compose run --rm -T dev lefthook run --no-auto-install %s "$@" </dev/null\n' "$hook" > "$HOOKS_DIR/$hook"; \
        chmod +x "$HOOKS_DIR/$hook"; \
    done

# Run standards checks inside the container
check:
    docker compose run --rm -T dev standards check </dev/null
    if git ls-files '*.swift' | grep -q .; then REPO_DIR="$REPO_DIR" sh -c '. ./bin/native-env.sh; git ls-files -z -- "*.swift" | xargs -0 swiftlint lint --strict --config .standards/config/swiftlint.yml'; fi
    if test -d Formula || test -d Casks; then brew style; fi

# Run formatters and fixers inside the container
fmt:
    docker compose run --rm -T dev standards fmt </dev/null

# Run pull-request conventions check inside the container
check-pr:
    docker compose run --rm -T dev standards check-pr </dev/null

# Open interactive shell inside the dev container
shell:
    docker compose run --rm dev bash

# Execute arbitrary command inside the dev container
exec *args:
    docker compose run --rm -T dev "$@" </dev/null

test *args:
    bin/test.sh "$@"

package *args:
    bin/package.sh "$@"

lint-go *args:
    docker compose run --rm -T dev bin/lint-go.sh "$@" </dev/null
