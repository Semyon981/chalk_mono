-- +up
CREATE TABLE IF NOT EXISTS "accounts" (
  "id" SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS "courses" (
  "id" SERIAL PRIMARY KEY,
  "account_id" BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS "idx_courses__account_id" ON "courses" ("account_id");

ALTER TABLE "courses" ADD CONSTRAINT "fk_courses__account_id" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON DELETE CASCADE;

CREATE TABLE IF NOT EXISTS "sections" (
  "id" SERIAL PRIMARY KEY,
  "course_id" BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS "idx_sections__course_id" ON "sections" ("course_id");

ALTER TABLE "sections" ADD CONSTRAINT "fk_sections__course_id" FOREIGN KEY ("course_id") REFERENCES "courses" ("id") ON DELETE CASCADE;

CREATE TABLE IF NOT EXISTS "blocks" (
  "id" SERIAL PRIMARY KEY,
  "section_id" BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS "idx_blocks__section_id" ON "blocks" ("section_id");

ALTER TABLE "blocks" ADD CONSTRAINT "fk_blocks__section_id" FOREIGN KEY ("section_id") REFERENCES "sections" ("id") ON DELETE CASCADE;

CREATE TABLE IF NOT EXISTS "files" (
  "id" SERIAL PRIMARY KEY,
  "block_id" BIGINT
);

CREATE INDEX IF NOT EXISTS "idx_files__block_id" ON "files" ("block_id");

ALTER TABLE "files" ADD CONSTRAINT "fk_files__block_id" FOREIGN KEY ("block_id") REFERENCES "blocks" ("id") ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS "users" (
  "id" SERIAL PRIMARY KEY,
  "name" TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS "accounts_users" (
  "id" SERIAL PRIMARY KEY,
  "user_id" BIGINT NOT NULL,
  "account_id" BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS "idx_accounts_users__account_id" ON "accounts_users" ("account_id");

CREATE INDEX IF NOT EXISTS "idx_accounts_users__user_id" ON "accounts_users" ("user_id");

ALTER TABLE "accounts_users" ADD CONSTRAINT "fk_accounts_users__account_id" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON DELETE CASCADE;

ALTER TABLE "accounts_users" ADD CONSTRAINT "fk_accounts_users__user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE;

CREATE TABLE IF NOT EXISTS "credentials" (
  "id" SERIAL PRIMARY KEY,
  "user_id" BIGINT NOT NULL,
  "email" TEXT NOT NULL,
  "hashpass" TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS "idx_credentials__user_id" ON "credentials" ("user_id");

ALTER TABLE "credentials" ADD CONSTRAINT "fk_credentials__user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id");

-- +down
DROP TABLE IF EXISTS "accounts";
DROP TABLE IF EXISTS "courses";
DROP TABLE IF EXISTS "sections";
DROP TABLE IF EXISTS "blocks";
DROP TABLE IF EXISTS "files";
DROP TABLE IF EXISTS "users";
DROP TABLE IF EXISTS "accounts_users";
DROP TABLE IF EXISTS "credentials";