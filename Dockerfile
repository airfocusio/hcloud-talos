FROM scratch
ENTRYPOINT ["/bin/hcloud-talos"]
COPY --from=ghcr.io/talos-systems/talosctl:v0.14.2 /talosctl /bin/talosctl
COPY hcloud-talos /bin/hcloud-talos
WORKDIR /workdir
