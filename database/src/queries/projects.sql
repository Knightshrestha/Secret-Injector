-- name: CreateProject :one
INSERT INTO
    project_list (id, name, description)
VALUES
    (
        sqlc.arg ('id'),
        sqlc.arg ('name'),
        sqlc.narg ('description')
    ) RETURNING *;

-- name: GetProjectByID :one
SELECT
    *
FROM
    project_list
WHERE
    id = sqlc.arg ('id');

-- name: GetAllProjects :many
SELECT
    *
FROM
    project_list;

-- name: UpdateProject :one
UPDATE project_list
SET
    name = COALESCE(sqlc.narg ('name'), name),
    description = COALESCE(sqlc.narg ('description'), description),
    updated_at = CURRENT_TIMESTAMP
WHERE
    id = sqlc.arg ('id') RETURNING *;

-- name: DeleteProject :exec
DELETE FROM project_list
WHERE
    id = sqlc.arg ('id');