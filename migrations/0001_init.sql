-- +up

-- users ---------------
CREATE TABLE IF NOT EXISTS "users" (
  "id" SERIAL PRIMARY KEY,
  "name" TEXT NOT NULL,
  "email" TEXT UNIQUE NOT NULL,
  "hashpass" TEXT NOT NULL
);


-- sessions ---------------
CREATE TABLE IF NOT EXISTS "sessions" (
  "id" SERIAL PRIMARY KEY,
  "user_id" BIGINT NOT NULL,
  "access_token" TEXT UNIQUE NOT NULL,
  "refresh_token" TEXT UNIQUE NOT NULL,
  "access_expires" TIMESTAMP NOT NULL,
  "refresh_expires" TIMESTAMP NOT NULL,
  "issued" TIMESTAMP NOT NULL,
  CONSTRAINT "fk_sessions__user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS "idx_sessions__user_id" ON "sessions" ("user_id");

-- accounts ---------------
CREATE TABLE IF NOT EXISTS "accounts" (
  "id" SERIAL PRIMARY KEY,
  "name" TEXT UNIQUE NOT NULL
);
CREATE TABLE IF NOT EXISTS "account_members" (
  "user_id" BIGINT NOT NULL,
  "account_id" BIGINT NOT NULL,
  "role" TEXT NOT NULL,
  PRIMARY KEY ("user_id", "account_id"),
  CONSTRAINT "fk_account_members__account_id" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON DELETE CASCADE,
  CONSTRAINT "fk_account_members__user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS "idx_account_members__account_id" ON "account_members" ("account_id");
CREATE INDEX IF NOT EXISTS "idx_account_members__user_id" ON "account_members" ("user_id");


-- files ---------------
CREATE TABLE IF NOT EXISTS "files" (
  "id" SERIAL PRIMARY KEY,
  "uploader_user_id" BIGINT NOT NULL,
  "name" TEXT NOT NULL,
  "content_type" TEXT NOT NULL,
  "bucket" TEXT NOT NULL,
  "key" TEXT NOT NULL,
  "uploaded_at" TIMESTAMP NOT NULL,
  "size" BIGINT NOT NULL,
  CONSTRAINT "fk_files__uploader_user_id" FOREIGN KEY ("uploader_user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS "idx_files__uploader_user_id" ON "files" ("uploader_user_id");

-- courses ---------------
CREATE TABLE IF NOT EXISTS "courses" (
  "id" SERIAL PRIMARY KEY,
  "account_id" BIGINT NOT NULL,
  "name" TEXT NOT NULL,
  CONSTRAINT "fk_courses__account_id" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON DELETE CASCADE,
  UNIQUE("id", "account_id")
);
CREATE INDEX IF NOT EXISTS "idx_courses__account_id" ON "courses" ("account_id");

-- CREATE TABLE "course_participants" (
--   "user_id" BIGINT NOT NULL,
--   "course_id" BIGINT NOT NULL,
--   PRIMARY KEY ("user_id", "course_id"),
--   CONSTRAINT "fk_course_participants__course_id" FOREIGN KEY ("course_id") REFERENCES "courses" ("id") ON DELETE CASCADE
--   CONSTRAINT "fk_course_participants__user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
-- );
-- CREATE INDEX "idx_course_participants__course_id" ON "course_participants" ("course_id");
-- CREATE INDEX "idx_course_participants__user_id" ON "course_participants" ("user_id");

CREATE TABLE IF NOT EXISTS "course_participants" (
  "user_id" BIGINT NOT NULL,
  "account_id" BIGINT NOT NULL,
  "course_id" BIGINT NOT NULL,
  PRIMARY KEY ("user_id", "course_id"),
  CONSTRAINT "fk_course_participants__user_id_account_id" FOREIGN KEY ("user_id", "account_id") REFERENCES account_members ("user_id", "account_id") ON DELETE CASCADE,
  CONSTRAINT "fk_course_participants__course_id_account_id" FOREIGN KEY ("course_id", "account_id") REFERENCES courses ("id", "account_id") ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS "idx_course_participants__course_id" ON "course_participants" ("course_id");
CREATE INDEX IF NOT EXISTS "idx_course_participants__user_id" ON "course_participants" ("user_id");


-- Функция-триггер, которая заполнит NEW.account_id
CREATE OR REPLACE FUNCTION trg_set_cp_account_id()
  RETURNS TRIGGER
AS $$
BEGIN
  -- вытягиваем account_id из courses по course_id
  SELECT account_id
    INTO NEW.account_id
    FROM courses
   WHERE id = NEW.course_id
   LIMIT 1;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Собственно триггер, который срабатывает перед вставкой

DROP TRIGGER IF EXISTS before_insert_course_participants ON course_participants;
CREATE TRIGGER before_insert_course_participants
  BEFORE INSERT ON course_participants
  FOR EACH ROW
  WHEN (NEW.account_id IS NULL)  -- только если не указали явно
  EXECUTE FUNCTION trg_set_cp_account_id();


-- modules ---------------
CREATE TABLE IF NOT EXISTS "modules" (
  "id" SERIAL PRIMARY KEY,
  "course_id" BIGINT NOT NULL,
  "order_idx" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  CONSTRAINT "fk_modules__course_id" FOREIGN KEY ("course_id") REFERENCES "courses" ("id") ON DELETE CASCADE,
  UNIQUE ("course_id", "order_idx")
);
CREATE INDEX IF NOT EXISTS "idx_modules__course_id" ON "modules" ("course_id");

-- lessons ---------------
CREATE TABLE IF NOT EXISTS "lessons" (
  "id" SERIAL PRIMARY KEY,
  "module_id" BIGINT NOT NULL,
  "order_idx" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  CONSTRAINT "fk_lessons__module_id" FOREIGN KEY ("module_id") REFERENCES "modules" ("id") ON DELETE CASCADE,
  UNIQUE ("module_id", "order_idx")
);
CREATE INDEX IF NOT EXISTS "idx_lessons__module_id" ON "lessons" ("module_id");

-- blocks ---------------
CREATE TABLE IF NOT EXISTS "blocks" (
  "id" SERIAL PRIMARY KEY,
  "lesson_id" BIGINT,
  "order_idx" INTEGER NOT NULL,
  "type" TEXT NOT NULL CHECK ("type" IN ('video', 'text')),
  CONSTRAINT "fk_blocks__lesson_id" FOREIGN KEY ("lesson_id") REFERENCES "lessons" ("id") ON DELETE CASCADE,
  UNIQUE ("lesson_id", "order_idx")
);
CREATE INDEX IF NOT EXISTS "idx_blocks__lesson_id" ON "blocks" ("lesson_id");

-- text_blocks ---------------
CREATE TABLE IF NOT EXISTS "text_blocks" (
  "id" BIGINT PRIMARY KEY,
  "content" TEXT NOT NULL,
  CONSTRAINT "fk_text_blocks__id" FOREIGN KEY ("id") REFERENCES "blocks" ("id")
);

-- video_blocks ---------------
CREATE TABLE IF NOT EXISTS "video_blocks" (
  "id" BIGINT PRIMARY KEY,
  "file_id" BIGINT NOT NULL,
  CONSTRAINT "fk_video_blocks__id" FOREIGN KEY ("id") REFERENCES "blocks" ("id"),
  CONSTRAINT "fk_video_blocks__file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id")
);
CREATE INDEX IF NOT EXISTS "idx_video_blocks__file_id" ON "video_blocks" ("file_id");

-- +down
DROP TABLE IF EXISTS "video_blocks";
DROP TABLE IF EXISTS "text_blocks";
DROP TABLE IF EXISTS "blocks";
DROP TABLE IF EXISTS "lessons";
DROP TABLE IF EXISTS "modules";
DROP TRIGGER IF EXISTS before_insert_course_participants
  ON course_participants;
DROP FUNCTION IF EXISTS trg_set_cp_account_id();
DROP TABLE IF EXISTS "course_participants";
DROP TABLE IF EXISTS "courses";
DROP TABLE IF EXISTS "files";
DROP TABLE IF EXISTS "account_members";
DROP TABLE IF EXISTS "accounts";
DROP TABLE IF EXISTS "sessions";
DROP TABLE IF EXISTS "users";