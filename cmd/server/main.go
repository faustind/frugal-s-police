package main

import (
	"log"
	"net/http"
	"os"

	"faustind/aubipo/pkg/app"
)

func main() {
	bot, err := app.NewKitchenSink(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
		os.Getenv("APP_BASE_URL"),
	)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", bot.Callback)
	http.HandleFunc("/check-due-dates", bot.CheckDueDates)
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
}
