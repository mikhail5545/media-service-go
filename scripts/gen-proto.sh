#!/usr/bin/env bash

set -euo pipefail

if ! command -v buf&>/dev/null; then
  echo "install buf: https://docs.buf.build/installation"
  exit 2
fi

OUT_DIR=$(mktemp -d)
trap 'rm -rf "${OUT_DIR}"' EXIT

buf generate --template buf.gen.yaml --output "${OUT_DIR}"

rm -rf pb_tmp || true
mv "${OUT_DIR}" pb_tmp
echo "Generated pb in pb_tmp/"