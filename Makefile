db-up:
    tursodb ./server.db --sync-server 0.0.0.0:8080

dev/templ:
	templ generate --watch --proxy="http://localhost:9980" --open-browser=false -v --proxybind="0.0.0.0"

dev/server:
	air \
    --build.cmd "go build -o tmp/bin/main ./cmd/main.go" --build.bin "tmp/bin/main" --build.delay "300" \
    --build.entrypoint "./tmp/bin/main" \
    --build.exclude_dir "node_modules" \
    --build.include_ext "go" \
    --build.stop_on_error "false" \
    --misc.clean_on_exit "true"

dev/tailwind:
	tailwindcss -i ./static/css/input.css -o ./static/css/output.css --watch --minify

dev/sync_assets:
	air \
	--build.cmd "templ generate --notify-proxy" \
	--build.bin "true" \
	--build.delay "200" \
	--build.exclude_dir "" \
	--build.include_dir "static" \
	--build.include_ext "js,css"

dev:
	make -j4 dev/tailwind dev/templ dev/server  dev/sync_assets

build:
	go build -o tmp/bin/main ./cmd/main.go
