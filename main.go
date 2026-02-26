package main

import (
	"dermify-api/cmd"
	"log"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
