package runtime

import "github.com/katsuakikureyama/GoHP/internal/switchstring"

// SwitchString guards the selector before byte-based dispatch. PHP loose
// comparison agrees with byte equality on this admitted nonnumeric-string lane.
func SwitchString(value any) string {
	s, ok := value.(string)
	if !ok || !switchstring.Admitted(s) {
		panic("switch requires a nonnumeric string")
	}
	return s
}
