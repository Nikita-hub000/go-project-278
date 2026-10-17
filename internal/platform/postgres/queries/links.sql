-- name: GetAllLinks :many
SELECT id, original_url, short_name, short_url
FROM links
ORDER BY CASE WHEN sqlc.arg(sort_rule)::text = 'id ASC' THEN id END ASC,
         CASE WHEN sqlc.arg(sort_rule)::text = 'id DESC' THEN id END DESC,
         CASE WHEN sqlc.arg(sort_rule)::text = 'created_at ASC' THEN created_at END ASC,
         CASE WHEN sqlc.arg(sort_rule)::text = 'created_at DESC' THEN created_at END DESC,
         id ASC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);


-- name: GetLinkByID :one
SELECT id, original_url, short_name, short_url
FROM links
WHERE id = $1;

-- name: GetLinkByShortName :one
SELECT id, original_url, short_name, short_url
FROM links
WHERE short_name = $1;


-- name: AddLink :one
INSERT INTO links(original_url, short_name, short_url)
VALUES ($1, $2, $3)
RETURNING id, original_url, short_name, short_url, created_at;

-- name: GetLinkByShortUrl :one
SELECT id, original_url, short_name, short_url
FROM links
WHERE short_url = $1;


-- name: UpdateLink :one
UPDATE links
SET original_url = $1,
    short_name   = $2,
    short_url    = $3
WHERE id = $4
RETURNING id, original_url, short_name, short_url, created_at;


-- name: DeleteLink :execrows
DELETE
FROM links
WHERE id = $1;

-- name: CountLinks :one
SELECT count(*) FROM links;
