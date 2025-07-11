package probe

import "fmt"

// ErrInvalidMode is an error returned for invalid sensor mode.
type ErrInvalidMode struct {
	Mode Mode
}

func (e ErrInvalidMode) Error() string {
	return fmt.Sprintf("invalid probe mode: %v", e.Mode)
}
