# Super Tic-Tac-Toe

An attempt to implement "Super Tic-Tac-Toe" as a proof of concept for trying out the Ebitengine game engine.

How to build a web version:

```text
cp $(go env GOROOT)/misc/wasm/wasm_exec.js .
env GOOS=js GOARCH=wasm go build -o super-tic-tac-toe.wasm github.com/pdxiv/super-tic-tac-toe
```

To serve the page locally, just use `python3 -m http.server` and navigate to http://localhost:8000/super-tic-tac-toe.html
