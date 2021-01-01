-- create users table
CREATE TABLE users (
    user_id TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    enabled INTEGER NOT NULL DEFAULT 1,
    CONSTRAINT users_pkey PRIMARY KEY (user_id)
);

-- create authorities table
CREATE TABLE authorities (
    user_id TEXT NOT NULL,
    authority TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

-- create index for users
CREATE UNIQUE INDEX ix_users_user_id
    on users (user_id,username,password)

-- create index for authorities
CREATE UNIQUE INDEX ix_auth_user_id
    on authorities (user_id,authority);

--INSERT INTO users (username, password, enabled)
--    values ('user',
--    '$2a$10$8.UnVuG9HHgffUDAlk8qfOuVGkqRzgVymGe07xd00DMxs.AQubh4a',
--    1);

--INSERT INTO authorities (username, authority)
--    values ('user', 'ROLE_USER');