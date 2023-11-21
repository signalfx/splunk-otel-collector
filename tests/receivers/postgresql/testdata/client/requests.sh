#!/bin/bash
##################################################################################################
# This file will exercise a target POSTGRES_SERVER's test_db whose schema is defined in the
# corresponding server's initdb.d/db.sql.  It will create the necessary replication slots and make
# a series of random insertions, updates, and deletions without end to help generate values for
# produced metrics.
# As implemented, test_db size will grow without bound if left is running, so this script isn't
# currently suitable for extended demonstrations or soak test purposes.
##################################################################################################

echopsql() {
  set -x
  psql postgresql://test_user:test_password@${POSTGRES_SERVER:-localhost}:5432/test_db "$@"
  local r=$?
  { set +x; } 2>/dev/null
  return $r
}

# create replication slots for postgres_replication_state metrics
create_replication_slots() {
for query in \
  "SELECT * FROM pg_create_physical_replication_slot('some_physical_replication_slot');" \
  "SELECT * FROM pg_create_logical_replication_slot('some_logical_replication_slot', 'test_decoding');"; do
  while true; do
    if echopsql -c "${query}"; then
      echo "${query} successful"
      break
    fi
    sleep 1
  done
done
}

random_str() {
  tr -dc A-Za-z0-9 </dev/urandom | head -c 16
}

r_int() {
  tr -dc 0-9 </dev/urandom | head -c $1
}

random_int() {
  r_int 7
}

random_float() {
  a=$(r_int 7)
  b=$(r_int 3)
  echo "$a.$b"
}

modify_table_one() {
while true; do
  sleep .15
  s_one=$(random_str)
  s_two=$(random_str)
  echopsql -c "insert into test_schema.table_one values ( '$s_one', '$s_two', now(), now() );"
  s_three=$(random_str)
  echopsql -c "insert into test_schema.table_one values ( '$s_two', '$s_three', now(), now() );"
  s_four=$(random_str)
  echopsql -c "insert into test_schema.table_one values ( '$s_three', '$s_four', now(), now() );"

  echopsql -c "select * from test_schema.table_one where string_one='$s_one';"
  echopsql -c "select timestamp_one from test_schema.table_one where string_two='$s_two';"

  echopsql -c "update test_schema.table_one set string_one='$s_four' where string_one='$s_three'"
  echopsql -c "delete from test_schema.table_one where string_one='$s_four';"
done
}

modify_table_two() {
while true; do
  sleep .15
  i_one=$(random_int)
  i_two=$(random_int)
  f_one=$(random_float)
  f_two=$(random_float)
  echopsql -c "insert into test_schema.table_two values ( $i_one, $i_two, $f_one, $f_two );"
  i_three=$(random_int)
  f_three=$(random_float)
  echopsql -c "insert into test_schema.table_two values ( $i_two, $i_three, $f_two, $f_three );"
  i_four=$(random_int)
  f_four=$(random_float)
  echopsql -c "insert into test_schema.table_two values ( $i_three, $i_four, $f_three, $f_four );"

  echopsql -c "select * from test_schema.table_two where int_one=$i_one;"
  echopsql -c "select float_one from test_schema.table_two where int_two=$i_two;"

  echopsql -c "update test_schema.table_two set int_one=$i_four where int_one=$i_three"
  echopsql -c "delete from test_schema.table_two where int_one=$i_four;"
done
}

echo "Beginning psql requests"
create_replication_slots & modify_table_one & modify_table_two
