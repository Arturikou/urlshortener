package main

import (
	"log"
	"os"
)

func main() {
	// allowed inside main() of package main — no diagnostics
	log.Fatal("fatal in main")
	log.Fatalf("fatalf in main: %v", "err")
	log.Fatalln("fatalln in main")
	os.Exit(0)
}

func badExit() {
	os.Exit(1) // want `call to os.Exit outside main\(\) of package main`
}

func badFatal() {
	log.Fatal("fatal error") // want `call to log.Fatal outside main\(\) of package main`
}

func badFatalf() {
	log.Fatalf("fatal: %v", "err") // want `call to log.Fatalf outside main\(\) of package main`
}

func badFatalln() {
	log.Fatalln("fatal") // want `call to log.Fatalln outside main\(\) of package main`
}

func good() {
	// no forbidden calls
}
