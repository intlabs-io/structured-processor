CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
    full_name TEXT NOT NULL,
    phone TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT '2023-10-01 00:00:00',
    updated_at TIMESTAMPTZ DEFAULT '2023-10-01 00:00:00'
  );
  INSERT INTO users (full_name, phone, created_at, updated_at)
  VALUES ('John Doe', '123-456-7890', '2023-10-01 00:00:00', '2023-10-01 00:00:00');