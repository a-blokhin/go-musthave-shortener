package main

import (
	"log"
	"os"
)

// This file should NOT trigger warnings for log.Fatal and os.Exit in main function

func main() {
	// These should be allowed in main function of main package
	log.Fatal("this is allowed in main")
	log.Fatalf("this is allowed in main: %s", "error")
	log.Fatalln("this is allowed in main")
	os.Exit(0)
}
