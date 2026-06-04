CREATE TABLE IF NOT EXISTS metrics (
    id SERIAL PRIMARY KEY,
    mtype VARCHAR(255) NOT NULL,
    delta INTEGER,
    value DOUBLE PRECISION
);

CREATE INDEX IF NOT EXISTS idx_metrics_mtype ON metrics(mtype);