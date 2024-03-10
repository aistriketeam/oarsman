# oarsman

a program like `postman` but for people who swim in the command line

currently this is a quick and dirty curl command generator + runner that targets OpenAPI services that accept json payloads via POST.

# example usage

```
oarsman http://your-server.local/openapi.json  # if the URL does not end in .json, "openapi.json" is appended by default
```

this launches a fuzzy finder, then a tview input form to input arguments

[![asciicast](https://asciinema.org/a/ZEb8QfLs9e69hMsXbKQ4ZYFsO.svg)](https://asciinema.org/a/ZEb8QfLs9e69hMsXbKQ4ZYFsO)