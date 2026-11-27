-- +goose Up
CREATE INDEX idx_tasks_done ON tasks (done);

-- +goose Down
DROP INDEX idx_tasks_done;