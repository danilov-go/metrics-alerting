package main

import "os"

func main() {
	os.Exit(1) // want `обнаружен вызов os.Exit`
}
