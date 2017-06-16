#!/usr/bin/env bash

set -e

RESULT=$(gofmt -d .)

if [ -n "$RESULT" ]; then
    echo "$RESULT"
    exit 1
fi
