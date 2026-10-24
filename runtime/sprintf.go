package runtime

import (
	"strconv"
	"strings"

	"github.com/katsuakikureyama/GoHP/internal/sprintfsubset"
)

// SprintfZeroPad implements only the documented %0Ns string/integer subset.
// PHP string widths count bytes; negative integer signs are ordinary bytes.
func SprintfZeroPad(format, value any) any {
	f, ok := format.(string)
	if !ok {
		panic("sprintf format requires string")
	}
	width, ok := sprintfsubset.Width(f)
	if !ok {
		panic("sprintf format outside supported %0Ns subset (width 0..4096)")
	}
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case int64:
		s = strconv.FormatInt(v, 10)
	default:
		panic("sprintf value requires string or integer")
	}
	if len(s) >= width {
		return s
	}
	return strings.Repeat("0", width-len(s)) + s
}
