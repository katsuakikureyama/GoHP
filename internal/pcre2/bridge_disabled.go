//go:build !gohp_pcre2 || !cgo

package pcre2

// Match reports a typed unavailability error when the binary was not built
// with both the gohp_pcre2 tag and cgo. It intentionally does not fall back to
// another regexp implementation.
func Match(Request) (Result, error) {
	return Result{}, ErrUnavailable
}
