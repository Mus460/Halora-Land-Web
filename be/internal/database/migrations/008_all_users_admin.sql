-- For now every user is an admin (per-product decision). The statement is
-- idempotent: it re-runs on every boot and simply forces the current role.
UPDATE users SET role = 'ADMIN' WHERE role <> 'ADMIN';
