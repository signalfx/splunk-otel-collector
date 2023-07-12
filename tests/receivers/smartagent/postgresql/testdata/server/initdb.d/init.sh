#!/bin/bash
##################################################################################################
# This file will enable the required pg_stat_statements extension for a database named test_db to
# allow the postgresql monitor queries to return valid results to generate metrics.  Any desired
# functionality that should occur on server startup should be added here. It is intended to run at
# server startup or via the postgresql container's docker-entrypoint-initdb.d autorun
# functionality.  It does not create the test_db database and requires that to have been created
# earlier (e.g. by the container's POSTGRES_DB env var).
##################################################################################################
set -euo pipefail

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "test_db" -c "CREATE EXTENSION IF NOT EXISTS pg_stat_statements;"
