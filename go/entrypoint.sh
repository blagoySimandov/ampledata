#!/bin/sh
set -e

echo "Running database migrations..."
/bin/migrate up

echo "Starting server..."
exec /bin/server
