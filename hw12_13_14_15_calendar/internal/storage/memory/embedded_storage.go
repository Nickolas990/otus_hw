package memorystorage

import (
	"context"
	"fmt"
	"sync"
	"time"

	//nolint:depguard
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/config"
	//nolint:depguard
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/errs"
	//nolint:depguard
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/logger"
	//nolint:depguard
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/storage"
	//nolint:depguard
	"github.com/google/uuid"
)

type EmbeddedStorage struct {
	eventMap map[string]storage.Event
	log      logger.Logger
	mu       sync.RWMutex
}

func (s *EmbeddedStorage) Update(id string, updatedEvent storage.Event) (storage.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем, существует ли событие с таким ID для модификации
	if _, ok := s.eventMap[id]; !ok {
		return storage.Event{}, fmt.Errorf("event with ID %s not found", id)
	}

	// Получаем список всех событий, которые могут конфликтовать с обновленным событием (исключая само это событие)
	var conflictingEvents []storage.Event
	for _, event := range s.eventMap {
		if event.ID != id && (event.StartTime.Before(updatedEvent.EndTime) && event.EndTime.After(updatedEvent.StartTime)) {
			conflictingEvents = append(conflictingEvents, event)
		}
	}

	// Если есть конфликтующие события, возвращаем ошибку
	if len(conflictingEvents) > 0 {
		return storage.Event{}, errs.ErrConflict{
			Code:    1,
			Message: "event conflict",
			Events:  conflictingEvents,
		}
	}

	// Вносим изменения в событие
	updatedEvent.ID = id
	s.eventMap[id] = updatedEvent

	// Возвращаем обновленное событие
	return updatedEvent, nil
}

func (s *EmbeddedStorage) Get(id string) (storage.Event, error) {
	value, ok := s.eventMap[id]
	if !ok {
		return value, errs.ErrNotFound{
			Code:    1,
			Message: "event not found",
		}
	}
	return s.eventMap[id], nil
}

func (s *EmbeddedStorage) Connect(ctx context.Context, config config.Config) error {
	_ = ctx
	_ = config
	return nil
}

func (s *EmbeddedStorage) Close(ctx context.Context) error {
	_ = ctx
	return nil
}

func New(logger logger.Logger) *EmbeddedStorage {
	return &EmbeddedStorage{
		eventMap: make(map[string]storage.Event),
		log:      logger,
	}
}

func (s *EmbeddedStorage) Create(event storage.Event) (storage.Event, error) {
	var eventSlice []storage.Event

	s.mu.RLock()
	eventSlice, _ = s.EventListInInterval(event.StartTime, event.EndTime)
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Повторная проверка на конфликты после повторного захвата блокировки может быть необходима,
	// если между RUnlock и Lock могли произойти изменения.
	// Но это зависит от конкретной логики и требований к целостности данных.

	if len(eventSlice) != 0 {
		return storage.Event{}, errs.ErrConflict{
			Code:    1,
			Message: "event already exists",
			Events:  eventSlice,
		}
	}
	event.ID = uuid.New().String()
	s.eventMap[event.ID] = event
	return event, nil
}

func (s *EmbeddedStorage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.eventMap, id)
	return nil
}

func (s *EmbeddedStorage) EventListForDate(date time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	eventList := make([]storage.Event, 0)
	for _, event := range s.eventMap {
		if event.StartTime.Equal(date) {
			eventList = append(eventList, event)
		}
	}
	return eventList, nil
}

func (s *EmbeddedStorage) EventListForWeek(date time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	eventSlice := make([]storage.Event, 0)
	startOfDay := date.Truncate(24 * time.Hour)     // Определяем начало дня для даты date
	endOfWeek := startOfDay.Add(time.Hour * 24 * 7) // Конец недели

	for _, event := range s.eventMap {
		// Проверяем, что событие начинается в диапазоне от начала даты date до конца недели
		if (event.StartTime.After(startOfDay) || event.StartTime.Equal(startOfDay)) && event.StartTime.Before(endOfWeek) {
			eventSlice = append(eventSlice, event)
		}
	}
	return eventSlice, nil
}

func (s *EmbeddedStorage) EventListForMonth(date time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	eventSlice := make([]storage.Event, 0)
	startOfDay := date.Truncate(24 * time.Hour) // Определяем начало дня для даты date
	endOfMonth := startOfDay.AddDate(0, 1, 0)   // Конец месяца

	for _, event := range s.eventMap {
		// Проверяем, что событие начинается в диапазоне от начала даты date до конца месяца
		if (event.StartTime.After(startOfDay) || event.StartTime.Equal(startOfDay)) && event.StartTime.Before(endOfMonth) {
			eventSlice = append(eventSlice, event)
		}
	}
	return eventSlice, nil
}

func (s *EmbeddedStorage) EventListInInterval(start time.Time, end time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	eventSlice := make([]storage.Event, 0)
	for _, event := range s.eventMap {
		// Проверяем, пересекается ли событие с заданным интервалом
		if (event.StartTime.Before(end) && event.EndTime.After(start)) ||
			(event.StartTime.Equal(start) || event.StartTime.Equal(end)) {
			eventSlice = append(eventSlice, event)
		}
	}
	return eventSlice, nil
}

func (s *EmbeddedStorage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.eventMap = make(map[string]storage.Event)
}
