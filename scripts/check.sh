#!/bin/sh
set -eu
export CGO_ENABLED=0 GOTOOLCHAIN=local
# PHP and the Go toolchain must exist. Tests deliberately fail instead of skipping PHP.
go version
php --version
go vet ./...
go test -count=1 -v ./...
mkdir -p build
go build -o build/gohp ./cmd/gohp
build/gohp -in tests/fixtures/core.php -out build/core.go -report build/core.report.json -php-version "$(php -r 'echo PHP_MAJOR_VERSION.".".PHP_MINOR_VERSION;')"
go build -o build/core build/core.go
php tests/fixtures/core.php > build/php.stdout
build/core > build/go.stdout
cmp build/php.stdout build/go.stdout
# No generated file may be produced for rejected source.
rm -f build/rejected.go
if build/gohp -in tests/fixtures/unsupported.php -out build/rejected.go; then
  echo 'ERROR: unsupported source accepted' >&2; exit 1
fi
test ! -e build/rejected.go
printf '%s\n' 'PASS: compiled native Go output matches PHP; unsupported source rejected'
