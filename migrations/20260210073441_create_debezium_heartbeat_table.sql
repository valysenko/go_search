-- +goose Up
-- +goose StatementBegin
CREATE TABLE debezium_heartbeat (last_heartbeat_ts TIMESTAMPTZ DEFAULT NOW() PRIMARY KEY);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS debezium_heartbeat;
-- +goose StatementEnd
