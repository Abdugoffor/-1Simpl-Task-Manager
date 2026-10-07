CREATE INDEX IF NOT EXISTS idx_hududs_name_trgm
    ON hududs USING GIN (name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_hududs_is_active_id
    ON hududs (is_active, id DESC);
