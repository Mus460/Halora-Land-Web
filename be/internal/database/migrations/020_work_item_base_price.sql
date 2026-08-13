-- reference price of the AHSP master at snapshot time; shown only in the
-- work item's own section (RAB/rekapitulasi keep using unitPrice)
ALTER TABLE work_items ADD COLUMN IF NOT EXISTS "basePrice" DECIMAL(15,2);