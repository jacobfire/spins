-- name: GetSpinHistory :many
SELECT * FROM spin
WHERE email = $1 OR user_id = $2;

-- name: AddSpin :one
INSERT INTO spin (
    user_id, combination, email, created_at
) VALUES (
             $1, $2, $3, now()
         )
RETURNING *;
