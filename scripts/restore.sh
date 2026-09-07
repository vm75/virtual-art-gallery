#!/bin/sh
set -eu

if [ "$#" -ne 2 ]; then echo "usage: restore.sh BACKUP_FILE DATA_DIR" >&2; exit 2; fi
backup_file=$1
data_dir=$2
[ -f "$backup_file" ] || { echo "backup does not exist: $backup_file" >&2; exit 1; }
if [ -e "$data_dir" ] && [ "$(find "$data_dir" -mindepth 1 -print -quit)" ]; then echo "restore destination must be new or empty: $data_dir" >&2; exit 1; fi
mkdir -p "$data_dir"
tar -xzf "$backup_file" -C "$data_dir"
echo "restored data to $data_dir"
