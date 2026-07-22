package main

import (
	"issueweave/cmd"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cmd.Execute()
}
