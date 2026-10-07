CREATE INDEX IF NOT EXISTS idx_categorys_name_trgm
    ON categorys USING GIN (name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_categorys_slug_trgm
    ON categorys USING GIN (slug gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_categorys_description_trgm
    ON categorys USING GIN (description gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_categorys_is_active_id
    ON categorys (is_active, id DESC);
