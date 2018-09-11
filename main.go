package main

import (
	"log"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lshortfile)
	log.SetOutput(os.Stdout)
	log.Println("starting")
}
