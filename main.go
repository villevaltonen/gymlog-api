package main

import (
	"os"

	"github.com/villevaltonen/gymlog-go/app"
)

func main() {
	app.CheckEnvVariableExists("JWT_KEY")
	a := app.Server{}
	a.Initialize(
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_HOST"))

	a.Run(":8010")

}
