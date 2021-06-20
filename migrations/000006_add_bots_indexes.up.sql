CREATE INDEX IF NOT EXISTS bot_name_idx ON bots USING GIN (to_tsvector('simple', bot_name));
CREATE INDEX IF NOT EXISTS bot_token_idx ON bots USING GIN (to_tsvector('simple', bot_token));
