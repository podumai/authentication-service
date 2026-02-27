CREATE SCHEMA bs;

CREATE TYPE role_title AS ENUM ('user', 'admin');
CREATE TYPE user_status AS ENUM ('active', 'inactive');

CREATE TABLE bs.roles (
  id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  title role_title NOT NULL,
  permissions TEXT[] NOT NULL
);

INSERT INTO bs.roles (title, permissions) VALUES ('user','{"read", "write"}'), ('admin', '{"read", "write"}');

CREATE TABLE bs.users (
  id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  role INTEGER NOT NULL,
  status user_status NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  password BYTEA NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT(CURRENT_TIMESTAMP),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT(CURRENT_TIMESTAMP),
  CONSTRAINT users_role_fk FOREIGN KEY (role) REFERENCES bs.roles (id),
  CONSTRAINT users_email_check CHECK(email ~* '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$')
);

CREATE USER backend WITH NOSUPERUSER
                         NOCREATEDB
                         NOCREATEROLE
                         NOINHERIT
                         LOGIN
                         NOREPLICATION
                         PASSWORD 'backend';

GRANT USAGE ON SCHEMA bs TO backend GRANTED BY postgres;
GRANT SELECT, INSERT, UPDATE, DELETE ON bs.users TO backend GRANTED BY postgres;
GRANT SELECT ON bs.roles TO backend GRANTED BY postgres;

CREATE TABLE bs.tokens (
  id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id INTEGER NOT NULL,
  refresh_token TEXT NOT NULL,
  device_name TEXT,
  ip_address INET,
  CONSTRAINT tokens_user_id_fk FOREIGN KEY (user_id) REFERENCES bs.users (id) ON UPDATE CASCADE ON DELETE CASCADE
);

GRANT SELECT, INSERT, UPDATE, DELETE ON bs.tokens TO backend GRANTED BY postgres;