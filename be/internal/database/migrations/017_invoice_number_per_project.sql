-- invoices: number uniqueness must be per project (numbering is per project+month)
ALTER TABLE invoices DROP CONSTRAINT IF EXISTS invoice_nomor_key;
ALTER TABLE invoices ADD CONSTRAINT invoices_project_number_key UNIQUE ("projectId", number);
