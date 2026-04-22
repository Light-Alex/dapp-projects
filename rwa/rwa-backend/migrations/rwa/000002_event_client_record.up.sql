-- Create event_client_record table for tracking processed blocks and events
CREATE TABLE IF NOT EXISTS event_client_record (
    chain_id BIGINT PRIMARY KEY,
    last_block BIGINT NOT NULL DEFAULT 0,
    last_event_id BIGINT NOT NULL DEFAULT 0,
    update_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_event_client_record_chain_id ON event_client_record (chain_id);

