-- +goose Up
-- Early deployments applied version 1 before short_url was added to that file.
ALTER TABLE links ADD COLUMN IF NOT EXISTS short_url TEXT;

UPDATE links
SET short_url = '/r/' || short_name
WHERE short_url IS NULL;

ALTER TABLE links ALTER COLUMN short_url SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS links_short_url_key ON links (short_url);

-- +goose Down
-- short_url is part of the baseline schema. Keep it and its data on rollback.
SELECT 1;
