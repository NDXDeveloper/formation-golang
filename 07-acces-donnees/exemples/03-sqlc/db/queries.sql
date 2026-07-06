-- name: GetAuthor :one
SELECT id, name, bio FROM authors WHERE id = $1;

-- name: CreateAuthor :one
INSERT INTO authors (name, bio) VALUES ($1, $2)
RETURNING id, name, bio;
