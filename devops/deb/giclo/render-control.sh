#!/bin/sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
CONTROL_TEMPLATE="${SCRIPT_DIR}/root/DEBIAN/control.tpl"
CONTROL_OUTPUT="${SCRIPT_DIR}/root/DEBIAN/control"

: "${PACKAGE_VERSION:?PACKAGE_VERSION is required}"

sed "s/@PACKAGE_VERSION@/${PACKAGE_VERSION}/g" "${CONTROL_TEMPLATE}" > "${CONTROL_OUTPUT}"
