CREATE TABLE IF NOT EXISTS letters (
    id TEXT PRIMARY KEY,
    original_id TEXT,
    subject TEXT,
    content TEXT,
    sender_email TEXT,
    sender_name TEXT,
    received_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    replied_at TIMESTAMP WITH TIME ZONE NULL,
    status TEXT DEFAULT 'new',
    ai_analysis JSONB NULL,
    category TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);
CREATE TABLE IF NOT EXISTS generated_replies (
    id TEXT PRIMARY KEY,
    letter_id TEXT REFERENCES letters(id) ON DELETE CASCADE,
    content TEXT,
    style TEXT,
    status TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);
-- simple index for queries
CREATE INDEX IF NOT EXISTS idx_letters_category ON letters(category);
CREATE INDEX IF NOT EXISTS idx_letters_status ON letters(status);