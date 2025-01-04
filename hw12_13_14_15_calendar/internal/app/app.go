package app

import (
	"context"
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/interfaces"
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/storage"
	"time"
)

type App struct {
	logger  interfaces.Logger
	storage interfaces.Storage
}

func New(logger interfaces.Logger, storage interfaces.Storage) *App {
	return &App{
		logger:  logger,
		storage: storage,
	}
}

func (a *App) CreateEvent(ctx context.Context, id, title string) error {
	return a.storage.Add(storage.Event{ID: id, Title: title})
}

func (a *App) DeleteEvent(ctx context.Context, id string) error {
	// TODO
	return a.storage.Delete(id)
}

func (a *App) ModifyEvent(ctx context.Context, id, title string) (storage.Event, error) {
	return a.storage.Modify(id, storage.Event{ID: id, Title: title})
}

func (a *App) GetEvent(ctx context.Context, id string) (storage.Event, error) {
	return a.storage.Get(id)
}

func (a *App) EventListForDate(ctx context.Context, date time.Time) ([]storage.Event, error) {
	return a.storage.EventListForDate(date)
}

func (a *App) EventListForWeek(ctx context.Context, date time.Time) ([]storage.Event, error) {
	return a.storage.EventListForWeek(date)
}

func (a *App) EventListForMonth(ctx context.Context, date time.Time) ([]storage.Event, error) {
	return a.storage.EventListForMonth(date)
}
