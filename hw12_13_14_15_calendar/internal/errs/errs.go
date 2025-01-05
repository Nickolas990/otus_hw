package errs

import (
	"fmt"

	//nolint:depguard
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/storage"
)

type ErrConflict struct {
	Code    int
	Message string
	Events  []storage.Event
}

func (e ErrConflict) Error() string {
	return fmt.Sprintf("code %d: %s", e.Code, e.Message)
}

type ErrNotFound struct {
	Code    int
	Message string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("code %d: %s", e.Code, e.Message)
}
