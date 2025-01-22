package main

import (
	"log"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("sleep")
	time.Sleep(time.Minute)
}
