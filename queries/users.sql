-- name: CreateUser :exec
INSERT INTO users (phone_number, password)
VALUES ($1, $2);

-- name: ExistUserByPhoneNumber :one
SELECT EXISTS(SELECT 1
              FROM users
              WHERE phone_number = $1) as exists;

-- name: GetUserByPhoneNumber :one
SELECT id,
       username,
       email,
       phone_number,
       role,
       password,
       created_at,
       updated_at
FROM users
WHERE phone_number = $1;


-- name: GetUserById :one
SELECT id,
       username,
       email,
       phone_number,
       role,
       password,
       created_at,
       updated_at
FROM users
WHERE id = $1;
