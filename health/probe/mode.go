package probe

import (
	"encoding/json"
	"log/slog"
	"sort"
	"strings"
)

// A Mode is a mode in which a sensor is run. This is a bitmask so sensors can run in multiple
// modes, for example startup and liveness.
type Mode uint8

// Supported sensor modes.
const (
	StartupMode Mode = 1 << iota
	ReadinessMode
	LivenessMode
)

// Common sensor modes
const (
	StartupLivenessMode = StartupMode | LivenessMode
	AllModes            = StartupMode | ReadinessMode | LivenessMode
)

var modeStrings = map[Mode]string{
	StartupMode:   "startup",
	LivenessMode:  "liveness",
	ReadinessMode: "readiness",
}

// ModeFromString returns a mode from the given string. If a valid mode does not exist
// returns a 0 mode and false, else the valid mode and true.
func ModeFromString(s string) (Mode, bool) {
	for k, v := range modeStrings {
		if strings.ToLower(s) == v {
			return k, true
		}
	}

	return Mode(0), false
}

// ValidMode returns a bool inddicating if the given mode is valid.
func ValidMode(mode Mode) bool {
	for k := range modeStrings {
		if k&mode == 0 {
			continue
		}

		return true
	}

	return false
}

func (m Mode) LogValue() slog.Value {
	var modes []string

	for mode, name := range modeStrings {
		if m&mode == 0 {
			continue
		}

		modes = append(modes, name)
	}

	return slog.AnyValue(modes)
}

func (m Mode) String() string {
	var v []string

	for mode, name := range modeStrings {
		if m&mode == 0 {
			continue
		}

		v = append(v, name)
	}

	sort.Strings(v)

	return strings.Join(v, ",")
}

// MarshalJSON marshals a sensor mode to a valid JSON string.
func (m Mode) MarshalJSON() ([]byte, error) {
	var v []string

	for mode, name := range modeStrings {
		if m&mode == 0 {
			continue
		}

		v = append(v, name)
	}

	if len(v) == 0 {
		return nil, ErrInvalidMode{Mode: m}
	}

	sort.Strings(v)

	return json.Marshal(v)
}
