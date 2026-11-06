# syntax=docker/dockerfile:1
FROM ghcr.io/rubio-enterprises/base:main AS dev

USER root
# Mise 2026.11.1 reads the system config's lock at /etc/mise/mise.lock
# Native dependency sidecars under image/.mise/ are part of the lock
COPY image/ /etc/mise/
RUN mv /etc/mise/mise.toml /etc/mise/config.toml \
 && chmod -R a+rX /etc/mise
RUN MISE_OVERRIDE_CONFIG_FILENAMES=/opt/standards/mise/standards.toml \
    node /opt/standards/image/locked-tools.mjs /etc/mise/config.toml /etc/mise/mise.lock \
 && tools="$(yq -p toml -oy -r '.tools | to_entries | .[] | select(.value.os == null or .value.os[] == "linux") | .key' /etc/mise/config.toml)" \
 && mise install --locked $tools \
 && mise reshim \
 && chmod -R a+rX /etc/mise

FROM dev AS ci
COPY --link --from=ghcr.io/rubio-enterprises/runner:main --chown=1000:1000 /home/runner /opt/actions-runner
WORKDIR /opt/actions-runner

# Same home and cache paths as the eventual runner, not root's throwaway caches
COPY --chown=1000:1000 . /workspaces/warmup
RUN cd /workspaces/warmup \
 && runuser -u vscode -- env HOME=/home/vscode bin/setup.sh \
 && rm -rf /workspaces/warmup
USER vscode
