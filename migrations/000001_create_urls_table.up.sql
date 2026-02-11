CREATE TABLE IF NOT EXISTS urls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    short_url VARCHAR(255) NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_urls_short_url ON urls(short_url);

CREATE INDEX IF NOT EXISTS idx_urls_original_url ON urls(original_url);