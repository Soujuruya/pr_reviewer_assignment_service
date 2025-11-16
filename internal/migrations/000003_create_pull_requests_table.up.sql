CREATE TABLE pull_requests (
    pull_request_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pull_request_name TEXT NOT NULL,
    author_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMPTZ DEFAULT now(),
    merged_at TIMESTAMPTZ
);
