help:
	@grep -B1 -h -E "^[a-zA-Z0-9_-]+\:([^\=]|$$)" $(MAKEFILE_LIST) \
    | grep -v -- -- \
    | grep -v '^help:' \
    | awk 'BEGIN {FS = ":"; ORS=""} \
        /^#/ {print $$0, "\n"} \
        /^[a-zA-Z0-9_-]+:/ {print "\033[34m", $$1, "\033[0m:\n"}'

# build the binary
build:
	go build -o oarsman

# dev in watch loop
dev:
	watchexec -w . -c -r -- make remote-example

# example against local server (you must start it)
remote-example:
	go run . http://localhost:8080/

# example against ./openapi.json (you must prepare it)
local-example:
	go run . ./openapi.json