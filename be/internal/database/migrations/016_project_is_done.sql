-- projects: add isDone (done/closed) status flag
ALTER TABLE projects ADD COLUMN IF NOT EXISTS "isDone" boolean NOT NULL DEFAULT false;
