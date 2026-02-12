DROP INDEX IF EXISTS idx_urls_original_url_unique;

CREATE INDEX IF NOT EXISTS idx_urls_original_url ON urls(original_url);