package memorystorage

import (
	"context"
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/business_errors"
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/config"
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/interfaces"
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	"sync"
	"time"
)

type EmbeddedStorage struct {
	// TODO
	eventMap map[string]storage.Event
	log      interfaces.Logger
	mu       sync.RWMutex
}

func (s *EmbeddedStorage) Modify(id string, event storage.Event) (storage.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	eventSlice, _ := s.EventListInInterval(event.StartTime, event.EndTime)
	if len(eventSlice) != 0 {
		return storage.Event{}, business_errors.ErrConflict{
			Code:    1,
			Message: "event conflict",
			Events:  eventSlice,
		}
	}
	s.eventMap[id] = event
	return event, nil
}

func (s *EmbeddedStorage) Get(id string) (storage.Event, error) {
	return s.eventMap[id], nil
}

func (s *EmbeddedStorage) Connect(ctx context.Context, config config.Config) error {
	return nil
}

func (s *EmbeddedStorage) Close(ctx context.Context) error {
	return nil
}

func New(logger interfaces.Logger) *EmbeddedStorage {
	return &EmbeddedStorage{
		eventMap: make(map[string]storage.Event),
		log:      logger,
	}
}

func (s *EmbeddedStorage) Add(event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	eventSlice, _ := s.EventListInInterval(event.StartTime, event.EndTime)
	if len(eventSlice) != 0 {
		return business_errors.ErrConflict{
			Code:    1,
			Message: "event already exists",
			Events:  eventSlice,
		}
	}
	event.ID = uuid.New().String()
	s.eventMap[event.ID] = event
	return nil
}

func (s *EmbeddedStorage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.eventMap, id)
	return nil
}

func (s *EmbeddedStorage) EventListForDate(date time.Time) ([]storage.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	eventList := make([]storage.Event, 0)
	for _, event := range s.eventMap {
		if event.StartTime.Equal(date) {
			eventList = append(eventList, event)
		}
	}
	return eventList, nil
}

func (s *EmbeddedStorage) EventListForWeek(date time.Time) ([]storage.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	eventSlice := make([]storage.Event, 0)
	for _, event := range s.eventMap {
		if event.StartTime.After(date) && event.StartTime.Before(date.Add(time.Hour*24*7)) {
			eventSlice = append(eventSlice, event)
		}
	}
	return eventSlice, nil
}

func (s *EmbeddedStorage) EventListForMonth(date time.Time) ([]storage.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	eventSlice := make([]storage.Event, 0)
	for _, event := range s.eventMap {
		if event.StartTime.After(date) && event.StartTime.Before(date.Add(time.Hour*24*30)) {
			eventSlice = append(eventSlice, event)
		}
	}
	return eventSlice, nil
}

func (s *EmbeddedStorage) EventListInInterval(start time.Time, end time.Time) ([]storage.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	eventSlice := make([]storage.Event, 0)
	for _, event := range s.eventMap {
		if event.StartTime.After(start) && event.StartTime.Before(end) {
			eventSlice = append(eventSlice, event)
		}
	}
	return eventSlice, nil
}
