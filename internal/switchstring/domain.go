// Package switchstring defines the nonnumeric-string switch admission boundary.
package switchstring

import "regexp"

// Numeric PHP8 strings permit ASCII whitespace, decimal fractions and exponents.
// Match syntax rather than parsing a float: overflowing numeric strings are
// numeric too. Everything outside this grammar compares by bytes in this lane.
var numeric = regexp.MustCompile(`^[\t\n\r\f\v ]*[+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE][+-]?[0-9]+)?[\t\n\r\f\v ]*$`)

func Admitted(value string) bool { return !numeric.MatchString(value) }
