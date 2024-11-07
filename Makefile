# run templ generation in watch mode to detect all .templ files and 
# re-create _templ.txt files on change, then send reload event to browser. 
# Default url: http://localhost:7331
# Define variables
GOPATH ?= $(shell go env GOPATH)
TEMPL = $(GOPATH)/bin/templ

live/templ:
	$(TEMPL) generate --watch --proxy="http://localhost:8080" --open-browser=false -v

# run air to detect any go file changes to re-build and re-run the server.
live/server:
	go run github.com/air-verse/air@v1.61.0 \
	--build.cmd "go build -o tmp/bin/main ./cmd/main.go" --build.bin "tmp/bin/main" --build.delay "100" \
	--build.exclude_dir "node_modules" \
	--build.include_ext "go" \
	--build.stop_on_error "false" \
	--misc.clean_on_exit true

# run tailwindcss to generate the styles.css bundle in watch mode.
live/tailwind:
	npx --yes tailwindcss -i ./input.css -o ./assets/styles.css --minify --watch

# start all 5 watch processes in parallel.
live: 
	make -j3 live/templ live/server live/tailwind
