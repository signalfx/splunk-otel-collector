-- This sql script is intended to be run via initdb or with the postgres container's
-- docker-entrypoint-initdb.d autorun functionality.  It will creeate the test schema
-- and test user to be exercised by the corresponding pg client's requests script to
-- assist in the postgresql monitor in generating metrics.

drop schema if exists test_schema cascade;
create schema test_schema;
set schema 'test_schema';

drop table if exists table_one;
create table table_one (
    string_one varchar(64) primary key,
    string_two varchar(64),
    timestamp_one timestamp not null,
    timestamp_two timestamp
);

drop table if exists table_two;
create table table_two (
    int_one integer primary key,
    int_two integer,
    float_one decimal(11,4) not null,
    float_two decimal(11,4)
);

drop role if exists test_user;
create role test_user with login password 'test_password';

alter user test_user with superuser;
grant pg_read_all_settings to test_user;
