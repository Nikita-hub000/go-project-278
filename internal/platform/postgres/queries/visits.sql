-- name: GetAllVisits :many
SELECT id, link_id, ip, status, user_agent, created_at
from visits
ORDER BY
    CASE WHEN sqlc.arg(sort_rule)::text = 'id ASC' THEN id END ASC,
    CASE WHEN sqlc.arg(sort_rule)::text = 'id DESC' THEN id END DESC,
    CASE WHEN sqlc.arg(sort_rule)::text = 'created_at ASC' THEN created_at END ASC,
    CASE WHEN sqlc.arg(sort_rule)::text = 'created_at DESC' THEN created_at END DESC,
    CASE WHEN sqlc.arg(sort_rule)::text = 'link_id ASC' THEN link_id END ASC,
    CASE WHEN sqlc.arg(sort_rule)::text = 'link_id DESC' THEN link_id END DESC,
    CASE WHEN sqlc.arg(sort_rule)::text = 'ip ASC' THEN ip END ASC,
    CASE WHEN sqlc.arg(sort_rule)::text = 'ip DESC' THEN ip END DESC,
    CASE WHEN sqlc.arg(sort_rule)::text = 'user_agent ASC' THEN user_agent END ASC,
    CASE WHEN sqlc.arg(sort_rule)::text = 'user_agent DESC' THEN user_agent END DESC,
    CASE WHEN sqlc.arg(sort_rule)::text = 'status ASC' THEN status END ASC,
    CASE WHEN sqlc.arg(sort_rule)::text = 'status DESC' THEN status END DESC,
    id ASC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: AddVisit :one
INSERT INTO visits(link_id, ip, status, user_agent)
VALUES ($1, $2, $3, $4)
RETURNING id, link_id, ip, status, user_agent, created_at;
