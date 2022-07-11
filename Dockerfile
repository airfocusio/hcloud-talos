FROM alpine:3.16 AS ca-certificates
RUN apk add ca-certificates

FROM scratch
ENTRYPOINT ["/bin/hcloud-talos"]
COPY --from=ca-certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=ghcr.io/siderolabs/talosctl:v1.1.0 /talosctl /bin/talosctl
COPY hcloud-talos /bin/hcloud-talos
WORKDIR /workdir
