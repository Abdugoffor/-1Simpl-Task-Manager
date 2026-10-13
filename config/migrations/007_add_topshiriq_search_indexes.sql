CREATE INDEX IF NOT EXISTS idx_topshiriqs_name_trgm
    ON topshiriqs USING GIN (name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_topshiriqs_description_trgm
    ON topshiriqs USING GIN (description gin_trgm_ops);
