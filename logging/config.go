package logging

import (
	"log"
	"os"
)

func Init() {
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lshortfile)
	log.SetOutput(os.Stdout)
}
