-- name: GetBalance :one
SELECT balance.balance FROM balance
WHERE email = $1 OR user_id = $2 LIMIT 1;

-- name: AddBalance :one
INSERT INTO balance (
    user_id, balance, email
) VALUES (
             $1, $2, $3
         )
RETURNING *;

-- name: UpdateBalance :exec
UPDATE balance
    SET balance = balance + $1
WHERE email = $2 OR user_id = $3;

-- name: UpdateWithNewValueBalance :exec
UPDATE balance
SET balance = $1
WHERE email = $2 OR user_id = $3;


-- name: SubBalance :exec
UPDATE balance
SET balance = balance - $1
WHERE balance > $1 AND email = $2 OR user_id = $3
RETURNING balance;