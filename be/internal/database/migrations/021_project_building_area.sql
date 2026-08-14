-- Luas bangunan (m²) for RAB header, captured at project creation.
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='projects' AND column_name='buildingArea')
    THEN ALTER TABLE "projects" ADD COLUMN "buildingArea" DECIMAL(10,2); END IF;
END $$;