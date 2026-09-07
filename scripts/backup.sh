#!/bin/sh
set -eu

if [ "$#" -ne 2 ]; then echo "usage: backup.sh DATA_DIR BACKUP_FILE" >&2; exit 2; fi
data_dir=$1
backup_file=$2
[ -d "$data_dir" ] || { echo "data directory does not exist: $data_dir" >&2; exit 1; }
[ ! -e "$backup_file" ] || { echo "backup already exists: $backup_file" >&2; exit 1; }
mkdir -p "$(dirname "$backup_file")"
tar -C "$data_dir" -czf "$backup_file" .
echo "backup written to $backup_file"
