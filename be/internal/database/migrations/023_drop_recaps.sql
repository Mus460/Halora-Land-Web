-- Margin/overhead removal: the legacy rekap/recaps table only stored the
-- per-project margin setting and imported BOQ divisions; nothing reads it.
-- Contract value is now synced from the RAB total instead.
DROP TABLE IF EXISTS "recaps";