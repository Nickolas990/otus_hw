package app

import (
	"context"
	"time"

	//nolint:depguard
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/logger"
	//nolint:depguard
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/storage"
)

type Application interface {
	CreateEvent(ctx context.Context, id, title string) (storage.Event, error)
	DeleteEvent(ctx context.Context, id string) error
	ModifyEvent(ctx context.Context, id, title string) (storage.Event, error)
	EventListForDate(ctx context.Context, date time.Time) ([]storage.Event, error)
	EventListForWeek(ctx context.Context, date time.Time) ([]storage.Event, error)
	EventListForMonth(ctx context.Context, date time.Time) ([]storage.Event, error)
}

type App struct {
	logger  logger.Logger
	storage storage.Storage
}

func New(logger logger.Logger, storage storage.Storage) *App {
	return &App{
		logger:  logger,
		storage: storage,
	}
}

func (a *App) CreateEvent(ctx context.Context, id, title string) (storage.Event, error) {
	select {
	case <-ctx.Done():
		return storage.Event{}, ctx.Err()
	default:
		return a.storage.Add(storage.Event{ID: id, Title: title})
	}
}

func (a *App) DeleteEvent(ctx context.Context, id string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return a.storage.Delete(id)
	}
}

func (a *App) ModifyEvent(ctx context.Context, id, title string) (storage.Event, error) {
	select {
	case <-ctx.Done():
		return storage.Event{}, ctx.Err()
	default:
		return a.storage.Modify(id, storage.Event{ID: id, Title: title})
	}
}

func (a *App) GetEvent(ctx context.Context, id string) (storage.Event, error) {
	select {
	case <-ctx.Done():
		return storage.Event{}, ctx.Err()
	default:
		return a.storage.Get(id)
	}
}

func (a *App) EventListForDate(ctx context.Context, date time.Time) ([]storage.Event, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return a.storage.EventListForDate(date)
	}
}

func (a *App) EventListForWeek(ctx context.Context, date time.Time) ([]storage.Event, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return a.storage.EventListForWeek(date)
	}
}

func (a *App) EventListForMonth(ctx context.Context, date time.Time) ([]storage.Event, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return a.storage.EventListForMonth(date)
	}
}
