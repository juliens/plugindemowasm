module github.com/traefik/plugindemowasm

go 1.22.3

require (
	github.com/http-wasm/http-wasm-guest-tinygo v0.3.0
	github.com/juliens/wasm-goexport v0.0.4
	github.com/stealthrocket/net v0.2.1
)

require github.com/tetratelabs/wazero v1.7.2 // indirect

replace github.com/http-wasm/http-wasm-guest-tinygo => github.com/juliens/http-wasm-guest-tinygo v0.0.0-20240602204949-9cdd64d990eb
