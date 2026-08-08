-- Progress history logs for work items (timestamped progress tracking)
CREATE TABLE IF NOT EXISTS work_item_progress_logs (
  id SERIAL PRIMARY KEY,
  "workItemId" INTEGER NOT NULL REFERENCES work_items(id) ON DELETE CASCADE,
  progress INTEGER NOT NULL DEFAULT 0,
  note TEXT,
  "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_progress_logs_workItem ON work_item_progress_logs ("workItemId");

-- Backfill one log entry per item that already has progress, timestamped with
-- its last update time so the S-curve has data from day one.
INSERT INTO work_item_progress_logs ("workItemId", progress, "createdAt")
SELECT id, progress, "updatedAt" FROM work_items
WHERE progress > 0
  AND NOT EXISTS (SELECT 1 FROM work_item_progress_logs WHERE "workItemId" = work_items.id);
