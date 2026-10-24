ARG PHP_IMAGE=php:8.3.36-cli@sha256:98d555ff97b3865140ad31032508c3ca089fa97a8944b14ca9dc03ea05d788cd
ARG GO_IMAGE=golang:1.23.12@sha256:60deed95d3888cc5e4d9ff8a10c54e5edc008c6ae3fba6187be6fb592e19e8c0
# Immutable Go toolchain and PHP OCI index references. PHP target selection is
# explicit in CI; add a later minor only after its own compatibility validation.
FROM ${GO_IMAGE} AS go
FROM ${PHP_IMAGE}
COPY --from=go /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:${PATH}" CGO_ENABLED=0 GOTOOLCHAIN=local GOCACHE=/tmp/gohp-go-cache GOMODCACHE=/tmp/gohp-mod-cache
WORKDIR /src
COPY . .
CMD ["sh", "scripts/check.sh"]
