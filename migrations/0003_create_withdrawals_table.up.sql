CREATE TABLE IF NOT EXISTS withdrawals (
                                           id SERIAL PRIMARY KEY,
                                           user_id INTEGER NOT NULL REFERENCES users(id),
                                           order_number VARCHAR(255) NOT NULL,
                                           sum NUMERIC(10,2) NOT NULL,
                                           processed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN withdrawals.id IS 'Уникальный идентификатор записи о списании';
COMMENT ON COLUMN withdrawals.user_id IS 'Идентификатор пользователя, выполнившего списание';
COMMENT ON COLUMN withdrawals.order_number IS 'Номер заказа, за который списываются баллы';
COMMENT ON COLUMN withdrawals.sum IS 'Количество баллов, списанных со счёта пользователя';
COMMENT ON COLUMN withdrawals.processed_at IS 'Время создания записи о списании баллов';