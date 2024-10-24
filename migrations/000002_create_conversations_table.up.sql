CREATE TABLE conversations (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    title TEXT NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id),
    history TEXT[] NOT NULL,
    version INTEGER NOT NULL DEFAULT 1
);
