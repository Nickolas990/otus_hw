-- +goose Up
-- SQL в этом секции будет выполнен при миграции вверх

CREATE TABLE IF NOT EXISTS events (
                                      id VARCHAR(36) PRIMARY KEY,
                                      title TEXT NOT NULL,
                                      description TEXT,
                                      start_time TIMESTAMP NOT NULL,
                                      end_time TIMESTAMP NOT NULL,
                                      user_id VARCHAR(36) NOT NULL,
                                      notification_time TIMESTAMP
);

-- +goose Down
-- SQL в этом секции будет выполнен при откате миграции

DROP TABLE IF EXISTS events;