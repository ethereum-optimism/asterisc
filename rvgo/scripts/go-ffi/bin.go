package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Must pass a subcommand")
	}
	switch os.Args[1] {
	case "diff":
		DiffTestUtils()
	default:
		log.Fatalf("Unrecognized subcommand: %s", os.Args[1])
	}
}
