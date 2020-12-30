package main

import (
	"os"

	"github.com/villevaltonen/gymlog-go/app"
)

func main() {
	a := app.Server{}
	a.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"))

	a.Run(":8010")

}
