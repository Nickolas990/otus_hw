package sqlstorage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/business_errors"
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/config"
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/interfaces"
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	_ "github.com/lib/pq" // Импортируем драйвер PostgreSQL
	"github.com/pressly/goose"
	"github.com/sirupsen/logrus"
	"time"
)

type PostgresStorage struct {
	db  *sql.DB
	log interfaces.Logger
}

const dbParamsKey = "dbParams"

func (s *PostgresStorage) Get(id string) (storage.Event, error) {
	const query = `SELECT id, title, description, start_time, end_time, user_id, notification_time FROM events WHERE id = $1`
	var event storage.Event

	row := s.db.QueryRow(query, id)
	err := row.Scan(&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime, &event.UserId, &event.NotificationTime)
	if err != nil {
		if err == sql.ErrNoRows {
			// Если событие не найдено, возвращаем ошибку
			return storage.Event{}, fmt.Errorf("no event found with id: %s", id)
		}
		// Возвращаем ошибку выполнения запроса
		return storage.Event{}, fmt.Errorf("error querying event with id %s: %w", id, err)
	}

	return event, nil
}

func (s *PostgresStorage) Add(event storage.Event) error {

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			logrus.Errorf("failed to rollback transaction: %s", err)
		}
	}(tx)

	event.ID = uuid.New().String()

	eventSlice, err := s.EventListInInterval(event.StartTime, event.EndTime)
	if len(eventSlice) != 0 {
		return business_errors.ErrConflict{
			Code:    1,
			Message: "event conflict",
			Events:  eventSlice,
		}
	}

	const query = `
    INSERT INTO events (id, title, start_time, end_time) 
    VALUES ($1, $2, $3, $4)`

	_, err = tx.Exec(query, event.ID, event.Title, event.StartTime, event.EndTime)
	if err != nil {
		return fmt.Errorf("failed to add event: %w", err)
	}

	return tx.Commit()
}

func (s *PostgresStorage) Delete(id string) error {
	// Начало транзакции
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Обязательное закрытие транзакции в конце выполнения метода
	defer func() {
		if p := recover(); p != nil {
			// Откат транзакции в случае паники
			tx.Rollback()
			panic(p) // Перевыброс паники
		} else if err != nil {
			// Откат транзакции в случае ошибки
			tx.Rollback()
		} else {
			// Попытка фиксации транзакции
			err = tx.Commit()
			if err != nil {
				err = fmt.Errorf("failed to commit transaction: %w", err)
			}
		}
	}()

	// Выполнение запроса в контексте транзакции
	const query = `DELETE FROM events WHERE id = $1`
	result, err := tx.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no event found with id: %s", id)
	}

	// Если дошли до этого момента без ошибок, err == nil,
	// и транзакция будет зафиксирована блоком defer выше.
	return nil
}

func (s *PostgresStorage) Modify(id string, event storage.Event) (storage.Event, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return storage.Event{}, fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // Возврат паники после отката
		} else if err != nil {
			_ = tx.Rollback() // Откат в случае ошибки
		} else {
			err = tx.Commit() // Попытка фиксации транзакции
		}
	}()

	// Использование EventListInInterval для проверки на конфликты
	events, err := s.EventListInInterval(event.StartTime, event.EndTime)
	if err != nil {
		return storage.Event{}, fmt.Errorf("failed to list events in interval: %w", err)
	}

	for _, existingEvent := range events {
		if existingEvent.ID != id {
			// Найден конфликтующий эвент, который не является модифицируемым эвентом
			return storage.Event{}, fmt.Errorf("event %s in time %s conflicts with an existing event", existingEvent.Title, existingEvent.StartTime)
		}
	}

	// Обновление события
	const updateQuery = `
UPDATE events 
SET title = $2, description = $3, start_time = $4, end_time = $5, user_id = $6, notification_time = $7 
WHERE id = $1
`
	if _, err := tx.Exec(updateQuery, id, event.Title, event.Description, event.StartTime, event.EndTime, event.UserId, event.NotificationTime); err != nil {
		return storage.Event{}, fmt.Errorf("failed to update event: %w", err)
	}

	// Возвращаем обновлённое событие
	return event, nil
}

func (s *PostgresStorage) EventListForDate(date time.Time) ([]storage.Event, error) {
	const query = `SELECT id, title, start_time, end_time FROM events WHERE DATE(start_time) = DATE($1)`
	rows, err := s.db.Query(query, date)
	if err != nil {
		return nil, fmt.Errorf("failed to list events for date: %w", err)
	}
	defer rows.Close()

	var events []storage.Event
	for rows.Next() {
		var e storage.Event
		if err := rows.Scan(&e.ID, &e.Title, &e.StartTime, &e.EndTime); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, e)
	}

	return events, nil
}

func (s *PostgresStorage) EventListForWeek(date time.Time) ([]storage.Event, error) {
	// Начало транзакции
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback() // Рекомендуется откатывать транзакцию, если она не будет зафиксирована

	// Вычисление начала и конца недели
	startOfWeek := date.Truncate(24 * time.Hour)            // Обрезать время суток
	offset := int(startOfWeek.Weekday()) - int(time.Monday) // Разница в днях от начала недели
	if offset > 0 {
		startOfWeek = startOfWeek.AddDate(0, 0, -offset) // Переместиться на начало недели
	} else if offset < 0 {
		startOfWeek = startOfWeek.AddDate(0, 0, -7+offset) // Для случаев, когда начало недели - в прошлом
	}
	endOfWeek := startOfWeek.AddDate(0, 0, 7) // Конец недели

	const query = `
SELECT id, title, description, start_time, end_time, user_id, notification_time 
FROM events 
WHERE start_time >= $1 AND start_time < $2`

	rows, err := tx.Query(query, startOfWeek, endOfWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to list events for week: %w", err)
	}
	defer rows.Close()

	var events []storage.Event
	for rows.Next() {
		var event storage.Event
		if err := rows.Scan(&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime, &event.UserId, &event.NotificationTime); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	// Проверка на ошибки при чтении результатов
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Фиксация транзакции не требуется, так как мы только читаем данные
	return events, nil
}

func (s *PostgresStorage) EventListForMonth(date time.Time) ([]storage.Event, error) {
	// Вычисление начала и конца месяца
	startOfMonth := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location()) // Первый день месяца
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)                    // Последний момент месяца

	const query = `
SELECT id, title, description, start_time, end_time, user_id, notification_time 
FROM events 
WHERE (start_time >= $1 AND start_time <= $2) OR
      (end_time >= $1 AND end_time <= $2) OR
      (start_time <= $1 AND end_time >= $2)`

	rows, err := s.db.Query(query, startOfMonth, endOfMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to list events for month: %w", err)
	}
	defer rows.Close()

	var events []storage.Event
	for rows.Next() {
		var event storage.Event
		if err := rows.Scan(&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime, &event.UserId, &event.NotificationTime); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	// Проверка на ошибки при чтении результатов
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (s *PostgresStorage) EventListInInterval(start, end time.Time) ([]storage.Event, error) {
	const query = `
SELECT id, title, description, start_time, end_time, user_id, notification_time 
FROM events 
WHERE (start_time >= $1 AND start_time <= $2) OR
      (end_time >= $1 AND end_time <= $2) OR
      (start_time <= $1 AND end_time >= $2)`

	rows, err := s.db.Query(query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to list events in interval: %w", err)
	}
	defer rows.Close()

	var events []storage.Event
	for rows.Next() {
		var event storage.Event
		if err := rows.Scan(&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime, &event.UserId, &event.NotificationTime); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	// Проверка на ошибки при чтении результатов
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func New(logger interfaces.Logger) *PostgresStorage {
	return &PostgresStorage{
		log: logger,
	}
}

func (s *PostgresStorage) Connect(ctx context.Context, cfg config.Config) error {

	params := cfg.DBParams
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		params.Host, params.Port, params.Username, params.Password, params.Database)
	db, err := sql.Open("postgres", dsn)

	if err != nil {
		return err
	}

	if err := db.PingContext(ctx); err != nil {
		err := db.Close()
		if err != nil {
			s.log.Errorf("failed to close db connection: %s", err)
			return err
		}
		return err
	}

	// Устанавливаем диалект для goose
	err = goose.SetDialect("postgres")
	if err != nil {
		s.log.Errorf("failed to set dialect: %s", err)
		return err
	}

	// Выполнение миграций
	migrationsDir := "hw12_13_14_15_calendar/migrations" // Укажите путь к директории с файлами миграции
	if err := goose.Up(db, migrationsDir); err != nil {
		err := db.Close()
		if err != nil {
			s.log.Errorf("failed to close db connection: %s", err)
			return err
		} // Закрываем соединение с БД в случае неудачной миграции
		return fmt.Errorf("goose failed to run migrations: %w", err)
	}

	s.db = db
	return nil
}

func (s *PostgresStorage) Close(ctx context.Context) error {
	if s.db != nil {
		err := s.db.Close()
		if err != nil {
			s.log.Errorf("failed to close db connection: %s", err)
			return err
		}
	}
	return nil
}
