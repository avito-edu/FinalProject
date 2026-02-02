CREATE TABLE IF NOT EXISTS questions (
  id          bigserial PRIMARY KEY,
  text        text NOT NULL,
  options     jsonb NOT NULL,
  correct_id  text NOT NULL,
  is_active   boolean NOT NULL DEFAULT true,
  created_at  timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE questions
  ADD CONSTRAINT questions_text_unique UNIQUE (text);

CREATE INDEX IF NOT EXISTS questions_active_idx ON questions (is_active);
