package memorystorage

import (
	"testing"
	"time"

	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/storage"
	"github.com/stretchr/testify/require"
)

var testDate = time.Date(2022, 10, 4, 12, 0, 0, 0, time.UTC)

func TestEmbeddedStorage(t *testing.T) {
	log := logger.New("debug")
	s := New(log)

	t.Cleanup(func() {
		s.Clear() // Очищаем хранилище после каждого теста
	})

	t.Run("Add", func(t *testing.T) {
		event := storage.Event{
			Title:     "Test Event",
			StartTime: testDate,
			EndTime:   testDate.Add(time.Hour),
		}

		storedEvent, err := s.Add(event)
		require.NoError(t, err)

		foundEvent, err := s.Get(storedEvent.ID)
		require.NoError(t, err)
		require.Equal(t, foundEvent, storedEvent)
	})

	t.Run("Delete", func(t *testing.T) {
		s.Clear()
		// Сначала добавляем событие, которое будем удалять
		event := storage.Event{
			Title:     "Test Event 2",
			StartTime: testDate,
			EndTime:   testDate.Add(time.Hour),
		}

		storedEvent, err := s.Add(event)
		require.NoError(t, err)

		err = s.Delete(storedEvent.ID)
		require.NoError(t, err)

		_, err = s.Get(storedEvent.ID)
		require.Error(t, err)
	})

	t.Run("Modify", func(t *testing.T) {
		s.Clear()
		// Добавляем событие для модификации
		event := storage.Event{
			Title:     "Event to Modify",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Hour),
		}
		storedEvent, err := s.Add(event)
		require.NoError(t, err)

		event.Title = "Modified Event"
		modifiedEvent, err := s.Modify(storedEvent.ID, event)
		require.NoError(t, err)

		foundEvent, err := s.Get(modifiedEvent.ID)
		require.NoError(t, err)
		require.Equal(t, "Modified Event", foundEvent.Title)
	})
}

func TestEmbeddedStorage_EventLists(t *testing.T) {
	log := logger.New("debug")
	s := New(log)

	t.Cleanup(func() {
		s.Clear() // Очищаем хранилище после каждого теста
	})

	t.Run("EventListForDate", func(t *testing.T) {
		s.Clear()                                  // Очищаем хранилище перед тестом
		addTestEvent(t, s, 0, time.Hour, testDate) // Добавляем событие на сегодня

		events, err := s.EventListForDate(testDate)
		require.NoError(t, err)
		require.Len(t, events, 1, "should find 1 event for the specific date")
	})

	t.Run("EventListForWeek", func(t *testing.T) {
		s.Clear()                                                // Очищаем хранилище перед тестом
		addTestEvent(t, s, 47*time.Hour, 48*time.Hour, testDate) // Через 2 дня
		addTestEvent(t, s, 0, 3*time.Hour, testDate)             // Сегодня

		events, err := s.EventListForWeek(testDate)
		require.NoError(t, err)
		require.Len(t, events, 2, "should find 2 events for this week")
	})

	t.Run("EventListForMonth", func(t *testing.T) {
		s.Clear()                                                                // Очищаем хранилище перед тестом
		addTestEvent(t, s, 15*24*time.Hour+time.Hour, 15*24*time.Hour, testDate) // 15 дней назад
		addTestEvent(t, s, 0, time.Hour, testDate)                               // Сегодня
		events, err := s.EventListForMonth(testDate)
		require.NoError(t, err)
		require.Len(t, events, 2, "should find 2 events for this month")
	})

	t.Run("EventListInInterval", func(t *testing.T) {
		s.Clear()                                                // Очищаем хранилище перед тестом
		addTestEvent(t, s, -2*time.Hour, -1*time.Hour, testDate) // Два часа назад
		addTestEvent(t, s, 0, time.Hour, testDate)               // Сегодня
		addTestEvent(t, s, 2*time.Hour, 3*time.Hour, testDate)   // Через два часа

		startInterval := testDate.Add(-1 * time.Hour)
		endInterval := testDate.Add(1 * time.Hour)
		events, err := s.EventListInInterval(startInterval, endInterval)
		require.NoError(t, err)
		require.Len(t, events, 1, "should find 1 event within the interval")
	})
}

func TestEmbeddedStorage_AddConflictEvent(t *testing.T) {
	log := logger.New("debug")
	s := New(log) // Предполагается, что New создает новый экземпляр EmbeddedStorage

	// Очищаем хранилище перед тестом
	t.Cleanup(func() {
		s.Clear()
	})

	// Добавляем первое событие
	firstEvent := storage.Event{
		Title:     "First Event",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(2 * time.Hour), // Длится 2 часа
	}

	_, err := s.Add(firstEvent)
	require.NoError(t, err)

	// Пытаемся добавить второе событие, которое пересекается по времени с первым
	conflictingEvent := storage.Event{
		Title:     "Conflicting Event",
		StartTime: firstEvent.StartTime.Add(1 * time.Hour), // Начинается через 1 час после начала первого события
		EndTime:   firstEvent.EndTime.Add(1 * time.Hour),   // Заканчивается через 1 час после окончания первого события
	}

	_, err = s.Add(conflictingEvent)
	require.Error(t, err, "expected error when adding a conflicting event")
}

func TestEmbeddedStorage_ModifyConflictEvent(t *testing.T) {
	log := logger.New("debug")
	s := New(log) // Предполагается, что New создает экземпляр хранилища

	// Очищаем хранилище перед тестом
	t.Cleanup(func() {
		s.Clear()
	})

	// Добавляем первое событие, которое не будет конфликтовать
	firstEvent := storage.Event{
		Title:     "First Event",
		StartTime: time.Now().Add(24 * time.Hour), // Начнется через 24 часа
		EndTime:   time.Now().Add(26 * time.Hour), // Закончится через 26 часов
	}
	firstEvent, err := s.Add(firstEvent)
	require.NoError(t, err)

	// Добавляем второе событие
	secondEvent := storage.Event{
		Title:     "Second Event",
		StartTime: time.Now().Add(48 * time.Hour), // Начнется через 48 часов
		EndTime:   time.Now().Add(50 * time.Hour), // Закончится через 50 часов
	}
	secondEvent, err = s.Add(secondEvent)
	require.NoError(t, err)

	// Пытаемся изменить второе событие так, чтобы оно конфликтовало с первым
	secondEvent.StartTime = firstEvent.StartTime.Add(time.Hour) // Должно вызвать конфликт
	secondEvent.EndTime = firstEvent.EndTime.Add(time.Hour)

	_, err = s.Modify(secondEvent.ID, secondEvent)
	require.Error(t, err, "expected error when modifying an event to a conflicting time")
}

func addTestEvent(t *testing.T, s *EmbeddedStorage, start, end time.Duration, baseDate time.Time) storage.Event {
	t.Helper()
	event := storage.Event{
		// ID и Title могут быть сгенерированы или указаны явно
		StartTime: baseDate.Add(start),
		EndTime:   baseDate.Add(end),
		// Другие поля события
	}
	addedEvent, err := s.Add(event)
	require.NoError(t, err)
	return addedEvent
}
