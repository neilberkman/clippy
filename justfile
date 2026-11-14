# host-first: macos
set positional-arguments
set shell := ["bash", "-euo", "pipefail", "-c"]

export RUBIO_BUILD_GITHUB_TOKEN := env('GH_TOKEN', env('GITHUB_TOKEN', ''))
export REPO_DIR := justfile_directory()
export COMPOSE_PROJECT_NAME := "repo-" + sha256(justfile_directory())
unexport GIT_COMMON_DIR
export RUBIO_GIT_COMMON_DIR := `env -u GIT_COMMON_DIR git rev-parse --path-format=absolute --git-common-dir`

_run := if path_exists('/opt/standards') == 'true' { '' } else { 'docker compose run --rm -T dev' }
interactive := if path_exists('/opt/standards') == 'true' { '' } else { 'docker compose run --rm dev' }

# Show available recipes
_default:
    @just --list

# Build dev container image
build:
    @if [ ! -d /opt/standards ]; then bin/build.sh; fi

# Build the dev image, sync dependencies into volumes, and install git hooks
[macos]
setup: build setup-macos

# Set up the macOS toolchain and shared config without requiring Docker
[macos]
setup-macos:
    bin/setup.sh
    @base="$(awk '/^FROM ghcr.io\/rubio-enterprises\/base:/ { print $2; exit }' Dockerfile)"; \
    test -n "$base"; mkdir -p .standards; \
    config="${STANDARDS_MACOS_CONFIG:-}"; \
    if test -n "$config"; then \
        source="$(dirname "$config")"; \
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
    @if command -v docker >/dev/null 2>&1; then docker compose run --rm -T dev /opt/standards/bin/install-hooks </dev/null; fi

# Run image Checks and macOS Checks on a local Mac
check:
    {{_run}} standards check </dev/null
    @if [ ! -d /opt/standards ]; then just check-macos; fi

# Run only macOS Checks; swept repositories keep their scoped local helper
[macos]
check-macos:
    bin/check-go.sh
    if test -f bin/macos-env.sh; then . ./bin/macos-env.sh; fi; if git ls-files '*.swift' | grep -q .; then if test -x bin/check-swift.sh; then bin/check-swift.sh; else swiftlint lint --strict --config .standards/config/swiftlint.yml; fi; fi
    if test -d Formula || test -d Casks; then brew style; fi

[macos]
test *args:
    bin/test.sh "$@"

[macos]
package *args:
    bin/package.sh "$@"

[macos]
lint-go *args:
    bin/lint-go.sh "$@"

# Run formatters and fixers inside the container
fmt:
    {{_run}} standards fmt </dev/null

# Run pull-request conventions check inside the container
check-pr:
    {{_run}} standards check-pr </dev/null

# Open interactive shell inside the dev container
shell:
    {{interactive}} bash

# Execute arbitrary command inside the dev container
exec *args:
    {{_run}} "$@" </dev/null
