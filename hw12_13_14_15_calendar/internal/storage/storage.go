package storage

import (
	"context"
	"time"

	//nolint:depguard
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/config"
)

type Storage interface {
	Add(event Event) (Event, error)
	Delete(id string) error
	Modify(id string, event Event) (Event, error)
	Get(id string) (Event, error)
	EventListForDate(date time.Time) ([]Event, error)
	EventListForWeek(date time.Time) ([]Event, error)
	EventListForMonth(date time.Time) ([]Event, error)
	EventListInInterval(start time.Time, end time.Time) ([]Event, error)
	Connect(ctx context.Context, config config.Config) error
	Close(ctx context.Context) error
}
