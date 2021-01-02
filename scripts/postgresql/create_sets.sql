CREATE TABLE IF NOT EXISTS sets
(
    id SERIAL,
    user_id TEXT NOT NULL,
	weight NUMERIC(10,2) NOT NULL DEFAULT 0.00,
	exercise TEXT NOT NULL,
	repetitions INTEGER,
	CONSTRAINT sets_pkey PRIMARY KEY (id)
);