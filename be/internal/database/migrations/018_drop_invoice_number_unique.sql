-- drop the old global-unique index on invoices.number (replaced by per-project
-- unique constraint invoices_project_number_key from migration 017)
DROP INDEX IF EXISTS invoice_nomor_key;
