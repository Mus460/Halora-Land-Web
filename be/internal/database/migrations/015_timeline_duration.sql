-- Timeline as structured duration: months + days (used by the S-curve planned line).
ALTER TABLE projects DROP COLUMN timeline;
ALTER TABLE projects ADD COLUMN "timelineMonths" integer NOT NULL DEFAULT 0;
ALTER TABLE projects ADD COLUMN "timelineDays" integer NOT NULL DEFAULT 0;
