templ:
	templ generate --watch --proxy="http://localhost:9980" --open-browser=false -v

db-up:
	rm data/dump*
	turso db shell app .dump > data/dump.sql
	cat data/dump.sql | sqlite3 data/dump.db
	turso dev --db-file data/dump.db

server:
	air \
    --build.cmd "go build -o tmp/bin/main ./cmd/main.go" \
    --build.bin "tmp/bin/main" \
    --build.delay "100" \
    --build.exclude_dir "node_modules" \
    --build.include_ext "go" \
    --build.stop_on_error "false" \
    --misc.clean_on_exit true

tailwind:
	tailwindcss -i ./static/css/input.css -o ./static/css/output.css --watch

dev:
	make -j 3 templ server tailwind
