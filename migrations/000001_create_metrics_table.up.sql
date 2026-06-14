CREATE TABLE IF NOT EXISTS metrics (
    id VARCHAR(255) PRIMARY KEY,
    mtype VARCHAR(255) NOT NULL,
    delta BIGINT,
    value DOUBLE PRECISION
);

CREATE INDEX IF NOT EXISTS idx_metrics_mtype ON metrics(mtype);