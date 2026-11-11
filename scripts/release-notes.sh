#!/bin/sh
# POSIX sh on purpose: bash 5.3 (Homebrew's default on macOS, and present on
# GitHub macOS runners) deadlocks on heredocs larger than ~512 bytes.
#
# Prints the GitHub release notes for one version: its CHANGELOG.md section
# under a "Changelog" heading. GoReleaser adds the release header and footer.
# Fails when the section is missing or empty, so a tag pushed before its
# changelog entry stops the release instead of publishing empty notes.
#
# Usage: scripts/release-notes.sh v1.2.3 [CHANGELOG.md]
set -eu

version=${1:?usage: release-notes.sh VERSION [CHANGELOG]}
version=${version#v}
changelog=${2:-CHANGELOG.md}

# The section runs from "## [VERSION]" to the next "## [" heading, without
# its leading blank lines; command substitution drops the trailing ones.
notes=$(awk -v heading="## [$version]" '
	index($0, heading) == 1 { found = 1; next }
	found && index($0, "## [") == 1 { exit }
	found && !started && $0 ~ /^[[:space:]]*$/ { next }
	found { started = 1; print }
' "$changelog")

if [ -z "$(printf '%s' "$notes" | tr -d '[:space:]')" ]; then
	echo "$changelog has no entry for $version; add a \"## [$version] - YYYY-MM-DD\" section" >&2
	exit 1
fi

printf '## Changelog\n\n%s\n' "$notes"
