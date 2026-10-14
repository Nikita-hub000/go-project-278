-- +goose Up
ALTER TABLE visits
    DROP CONSTRAINT visits_link_id_fkey,
    ADD CONSTRAINT visits_link_id_fkey
        FOREIGN KEY (link_id) REFERENCES links (id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE visits
    DROP CONSTRAINT visits_link_id_fkey,
    ADD CONSTRAINT visits_link_id_fkey
        FOREIGN KEY (link_id) REFERENCES links (id);
