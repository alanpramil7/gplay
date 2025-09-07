package main

import (
	"github.com/alanpramil7/gplay/cmd"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cmd.Execute()
}
