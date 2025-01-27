CREATE TABLE IF NOT EXISTS user_message_logs (
    user_id TEXT NOT NULL,
    message TEXT NOT NULL,
    stage TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
