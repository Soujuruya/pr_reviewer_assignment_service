CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT NOT NULL,
    team_name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);
