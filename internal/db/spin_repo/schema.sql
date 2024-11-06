CREATE TABLE IF NOT EXISTS spin(
    spin_id serial PRIMARY KEY,
    user_id INT REFERENCES users (user_id),
    combination VARCHAR(32),
    created_at timestamp,
    email VARCHAR (300) NOT NULL
);
