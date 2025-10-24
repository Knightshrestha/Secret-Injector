-- name: CreateSecret :one
INSERT INTO
    secret_list (id, project_id, key, value, description)
VALUES
    (
        sqlc.arg ('id'),
        sqlc.arg ('project_id'),
        sqlc.arg ('key'),
        sqlc.arg ('value'),
        sqlc.narg ('description')
    ) RETURNING *;

-- name: GetSecretByID :one
SELECT
    *
FROM
    secret_list
WHERE
    id = sqlc.arg ('id');

-- name: GetAllSecrets :many
SELECT
    *
FROM
    secret_list;

-- name: GetSecretsByProjectID :many
SELECT
    *
FROM
    secret_list
WHERE
    project_id = sqlc.arg ('project_id');

-- name: UpdateSecret :one
UPDATE secret_list
SET
    key = COALESCE(sqlc.narg ('key'), key),
    description = COALESCE(sqlc.narg ('description'), description),
    value = COALESCE(sqlc.narg ('value'), value),
    updated_at = CURRENT_TIMESTAMP
WHERE
    id = sqlc.arg ('id') RETURNING *;

-- name: DeleteSecret :exec
DELETE FROM secret_list
WHERE
    id = sqlc.arg ('id');

-- name: DeleteAllSecretsInProjects :exec
DELETE FROM secret_list
WHERE
    project_id = sqlc.arg ('project_id');