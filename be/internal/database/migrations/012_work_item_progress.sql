-- Progress tracking for work items (monitoring page)
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_items' AND column_name='progress')
    THEN ALTER TABLE "work_items" ADD COLUMN "progress" INTEGER NOT NULL DEFAULT 0; END IF;
END $$;
