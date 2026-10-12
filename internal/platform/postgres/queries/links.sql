-- name: GetAllLinks :many
SELECT id, original_url, short_name, short_url
FROM links
ORDER BY id
LIMIT $1 OFFSET $2;


-- name: GetLinkByID :one
SELECT id, original_url, short_name, short_url
FROM links
WHERE id = $1;


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
    short_name   = $2
WHERE id = $3
RETURNING id, original_url, short_name, short_url, created_at;


-- name: DeleteLink :execrows
DELETE
FROM links
WHERE id = $1;

