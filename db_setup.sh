sudo -u root bash -c "psql -h 127.0.0.1 -p 5432 -d fs3 < $(pwd)/fs3_db.sql"
