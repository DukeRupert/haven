-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS schedule_publications (
   id SERIAL PRIMARY KEY,
   created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
   updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
   facility_id INTEGER NOT NULL REFERENCES facilities(id) ON DELETE RESTRICT,
   published_through DATE NOT NULL,
   CONSTRAINT unique_facility_publication UNIQUE(facility_id)
);

CREATE INDEX idx_schedule_publications_facility ON schedule_publications(facility_id);
CREATE INDEX idx_schedule_publications_date ON schedule_publications(published_through);

CREATE TRIGGER update_schedule_publications_updated_at
   BEFORE UPDATE ON schedule_publications
   FOR EACH ROW
   EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_schedule_publications_updated_at ON schedule_publications;
DROP INDEX IF EXISTS idx_schedule_publications_facility;
DROP INDEX IF EXISTS idx_schedule_publications_date;
DROP TABLE IF EXISTS schedule_publications;
-- +goose StatementEnd