package interfaces

import (
	"context"
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/storage"
	"time"
)

type Application interface {
	CreateEvent(ctx context.Context, id, title string) error
	DeleteEvent(ctx context.Context, id string) error
	ModifyEvent(ctx context.Context, id, title string) (storage.Event, error)
	EventListForDate(ctx context.Context, date time.Time) ([]storage.Event, error)
	EventListForWeek(ctx context.Context, date time.Time) ([]storage.Event, error)
	EventListForMonth(ctx context.Context, date time.Time) ([]storage.Event, error)
}
