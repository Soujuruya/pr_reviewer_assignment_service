CREATE TABLE IF NOT EXISTS pull_request_reviewers (
    pull_request_id UUID NOT NULL,
    user_id UUID NOT NULL,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (pull_request_id, user_id),
    CONSTRAINT fk_prr_pr FOREIGN KEY (pull_request_id) REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    CONSTRAINT fk_prr_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_prr_user_id ON pull_request_reviewers(user_id);
