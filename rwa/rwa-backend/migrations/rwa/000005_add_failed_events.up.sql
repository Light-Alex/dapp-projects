CREATE TABLE IF NOT EXISTS failed_events (
    id BIGSERIAL PRIMARY KEY,
    client_order_id VARCHAR(255),
    event_type VARCHAR(50) NOT NULL,
    execution_id VARCHAR(255),
    event_data TEXT NOT NULL,
    error_message TEXT,
    source VARCHAR(50) NOT NULL DEFAULT 'alpaca',
    retry_count INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    resolved_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_failed_events_client_order_id ON failed_events (client_order_id);
CREATE INDEX idx_failed_events_status ON failed_events (status);
CREATE INDEX idx_failed_events_created_at ON failed_events (created_at);
