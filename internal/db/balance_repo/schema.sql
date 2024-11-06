CREATE TABLE IF NOT EXISTS balance(
    balance_id serial PRIMARY KEY,
    user_id INT REFERENCES users (user_id),
    balance FLOAT,
    email VARCHAR (300) NOT NULL
);
