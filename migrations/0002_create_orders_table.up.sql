CREATE TABLE IF NOT EXISTS orders (
                                      id SERIAL PRIMARY KEY,
                                      number VARCHAR(255) NOT NULL UNIQUE,
                                      user_id INTEGER NOT NULL REFERENCES users(id),
                                      status VARCHAR(50) NOT NULL,
                                      uploaded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                      accrual NUMERIC(10, 2) NOT NULL DEFAULT 0
);

COMMENT ON COLUMN orders.id IS 'Уникальный идентификатор заказа';
COMMENT ON COLUMN orders.number IS 'Номер заказа';
COMMENT ON COLUMN orders.user_id IS 'Идентификатор пользователя, загрузившего заказ';
COMMENT ON COLUMN orders.status IS 'Статус обработки заказа: NEW, PROCESSING, INVALID, PROCESSED';
COMMENT ON COLUMN orders.uploaded_at IS 'Время загрузки заказа';
COMMENT ON COLUMN orders.accrual IS 'Начисленные баллы для заказа';
