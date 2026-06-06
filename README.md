# oarsman

![](./logo.webp)

a program like `postman` but for people who swim in the command line

currently this is a quick and dirty curl command generator + runner that targets OpenAPI services that accept json payloads via POST.

# running

with [nix](https://nixos.org/) (flakes enabled), run directly without cloning:

```
nix run github:aistriketeam/oarsman -- http://your-server.local/openapi.json
```

or build with the included `Makefile` (`make build`).

# example usage

```
oarsman http://your-server.local/openapi.json  # if the URL does not end in .json, "openapi.json" is appended by default
```

this launches a fuzzy finder, then a tview input form to input arguments

[![asciicast](https://asciinema.org/a/653775.svg)](https://asciinema.org/a/653775)
