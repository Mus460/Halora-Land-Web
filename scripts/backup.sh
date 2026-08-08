#!/usr/bin/env bash
# Nightly Postgres backup for the production stack (VPS, /opt/halora).
#
# Install as a cron job (root):
#   0 2 * * * /opt/halora/scripts/backup.sh >> /var/log/halora-backup.log 2>&1
#
set -euo pipefail

STACK_DIR="${STACK_DIR:-/opt/halora}"
BACKUP_DIR="${BACKUP_DIR:-/root/backups}"
KEEP_DAYS="${KEEP_DAYS:-7}"

mkdir -p "$BACKUP_DIR"

echo "[$(date '+%F %T')] dumping halora DB..."
docker compose -f "$STACK_DIR/docker-compose.prod.yml" exec -T postgres \
  pg_dump -U postgres -d halora > "$BACKUP_DIR/halora-$(date '+%F-%H%M').sql"

echo "[$(date '+%F %T')] pruning backups older than ${KEEP_DAYS} days..."
find "$BACKUP_DIR" -name 'halora-*.sql' -mtime +"$KEEP_DAYS" -delete

echo "[$(date '+%F %T')] backup done."
