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
setup: build setup-native

# Run standards checks inside the container
check:
    docker compose run --rm -T dev standards check </dev/null
    @just check-native

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
    bin/lint-go.sh "$@"

setup-native:
    bin/setup.sh
    @base="$(awk '/^FROM ghcr.io\/rubio-enterprises\/base:/ { print $2; exit }' Dockerfile)"; \
    test -n "$base"; mkdir -p .standards; \
    if test -n "${STANDARDS_NATIVE_CONFIG:-}"; then \
        source="$(dirname "$STANDARDS_NATIVE_CONFIG")"; \
        test -f "$source/config/swiftlint.yml"; test -f "$source/mise/scope.sh"; test -f "$source/lefthook/standards.yml"; \
        for dir in config mise lefthook; do rm -rf ".standards/$dir"; cp -R "$source/$dir" ".standards/$dir"; done; \
    elif command -v docker >/dev/null 2>&1; then \
        cid="$(docker create "$base")"; trap 'docker rm "$cid" >/dev/null' EXIT; \
        docker cp "$cid:/opt/standards/." .standards/; \
    else \
        version="${base##*:}"; case "$version" in main) version=latest ;; v*) version="${version#v}" ;; esac; \
        tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT; \
        NODE_AUTH_TOKEN="${NODE_AUTH_TOKEN:-${GH_TOKEN:-${GITHUB_TOKEN:-}}}" env \
            'npm_config_@rubio-enterprises:registry=https://npm.pkg.github.com' \
            'npm_config_//npm.pkg.github.com/:_authToken=${NODE_AUTH_TOKEN}' \
            npm pack "@rubio-enterprises/standards@$version" --ignore-scripts --pack-destination "$tmp" >/dev/null; \
        tar -xzf "$tmp/"*.tgz -C "$tmp" --strip-components=1 package/config package/mise package/lefthook; \
        test -f "$tmp/config/swiftlint.yml"; test -f "$tmp/mise/scope.sh"; test -f "$tmp/lefthook/standards.yml"; \
        for dir in config mise lefthook; do rm -rf ".standards/$dir"; mv "$tmp/$dir" ".standards/$dir"; done; \
    fi
    @if ! command -v docker >/dev/null 2>&1; then exit 0; fi; \
    HOOKS_DIR="$(git rev-parse --git-path hooks)"; \
    mkdir -p "$HOOKS_DIR"; \
    rm -f "$HOOKS_DIR/post-checkout" "$HOOKS_DIR/post-merge" "$HOOKS_DIR/post-rewrite"; \
    for hook in pre-commit commit-msg pre-push; do \
        printf '#!/bin/sh\nset -eu\nREPO_DIR="$(git rev-parse --show-toplevel)"\nexport REPO_DIR\nexport GIT_COMMON_DIR="$(git -C "$REPO_DIR" rev-parse --path-format=absolute --git-common-dir)"\nexport COMPOSE_PROJECT_NAME="repo-$(printf %%s "$REPO_DIR" | shasum -a 256 | cut -d " " -f 1)"\ncd "$REPO_DIR" && exec docker compose run --rm -T dev lefthook run --no-auto-install %s "$@" </dev/null\n' "$hook" > "$HOOKS_DIR/$hook"; \
        chmod +x "$HOOKS_DIR/$hook"; \
    done

check-native:
    if test -f bin/native-env.sh; then . ./bin/native-env.sh; fi; if git ls-files '*.swift' | grep . >/dev/null; then if test -x bin/check-swift.sh; then bin/check-swift.sh; else swiftlint lint --strict --config .standards/config/swiftlint.yml; fi; fi
    if test -d Formula || test -d Casks; then brew style; fi
