#! /bin/bash

rm data/local.db*
rm data/dump*

turso db shell app .dump > data/dump.sql
cat data/dump.sql | sqlite3 data/dump.db
touch data/local.db

turso dev --db-file data/dump.db
