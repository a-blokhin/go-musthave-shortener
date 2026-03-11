package a

import (
	"log"
	"os"
)

// This file should trigger warnings for panic, log.Fatal, and os.Exit

func badPanic() {
	panic("this should be reported") // want "panic should not be used in production code"
}

func badLogFatal() {
	log.Fatal("this should be reported") // want "log.Fatal should only be used in main function of main package"
}

func badLogFatalf() {
	log.Fatalf("this should be reported: %s", "error") // want "log.Fatal should only be used in main function of main package"
}

func badLogFatalln() {
	log.Fatalln("this should be reported") // want "log.Fatal should only be used in main function of main package"
}

func badOsExit() {
	os.Exit(1) // want "os.Exit should only be used in main function of main package"
}
