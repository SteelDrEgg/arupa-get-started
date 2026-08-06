//go:build wasip1

package main

import "github.com/SteelDrEgg/arupa-sdk/golang/wasm"

func main() {}

func init() {
	wasm.RegisterService(newService().runtime)
}
