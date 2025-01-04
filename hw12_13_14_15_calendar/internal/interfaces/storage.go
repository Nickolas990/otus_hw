package interfaces

import (
	"context"
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/config"
	event "github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/storage"
	"time"
)

type Storage interface {
	Add(event event.Event) error
	Delete(id string) error
	Modify(id string, event event.Event) (event.Event, error)
	Get(id string) (event.Event, error)
	EventListForDate(date time.Time) ([]event.Event, error)
	EventListForWeek(date time.Time) ([]event.Event, error)
	EventListForMonth(date time.Time) ([]event.Event, error)
	EventListInInterval(start time.Time, end time.Time) ([]event.Event, error)
	Connect(ctx context.Context, config config.Config) error
	Close(ctx context.Context) error
}
