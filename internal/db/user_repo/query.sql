-- name: GetUser :one
SELECT * FROM users
WHERE email = $1 AND password = $2 LIMIT 1;

-- name: InsertUser :one
INSERT INTO users (
    username, password, email
) VALUES (
             $1, $2, $3
         )
RETURNING *;
