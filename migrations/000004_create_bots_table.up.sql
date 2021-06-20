CREATE TABLE IF NOT EXISTS bots (
    id bigserial PRIMARY KEY,
    owner_id bigserial NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    bot_name text NOT NULL,
    bot_token text NOT NULL,
    bot_token_valid bool NOT NULL default FALSE,
    bot_isActive bool NOT NULL default FALSE,
    Constraint fk_user_id FOREIGN KEY(owner_id) references users(id)
);
