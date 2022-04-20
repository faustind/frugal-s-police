package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"faustind/aubipo/pkg/models"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func main() {
	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Setup HTTP Server for receiving requests from LINE platform
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		events, err := bot.ParseRequest(req)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}

		if len(events) == 0 {
			// Just the LINE platform checking if bot is alive
			w.WriteHeader(200)
			return
		}

		for _, event := range events {
			switch event.Type {
			case linebot.EventTypeFollow:
				// save user to db
				// reply with "Set your monthly budget by sending `budget amount"
				if event.Source.Type != linebot.EventSourceTypeUser {
					log.Print("Follow event from non-user.")
					replyMsg := "Sorry, I can only talk to one user at a time!"
					if err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMsg)); err != nil {
						log.Print(err)
					}

				}

				log.Print("LINE EVENT: follow")

				var CreateUser = &models.User{}

				CreateUser.ID = event.Source.UserID
				user := CreateUser.CreateUser()

				log.Print(user)

				// send instruction to set budget
				replyMessage := "Got followed event"
				if err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMessage)); err != nil {
					log.Print(err)
				}

			case linebot.EventTypeUnfollow:
				// remove user and her subs from db
				if event.Source.Type != linebot.EventSourceTypeUser {
					log.Print("Follow event from non-user.")
				}

				log.Print("LINE EVENT: unfollow")

				userId := event.Source.UserID

				user := models.DeleteUser(userId)

				log.Print(user)

				replyMessage := "Bye!"
				log.Println(replyMessage)
				if err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMessage)); err != nil {
					log.Print(err)
				}

			case linebot.EventTypeMessage:
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					msg := strings.Fields(strings.ToLower(message.Text))

					if len(msg) == 2 && msg[0] == "yen" {
						// create/update user budget
						log.Print("YEN MSG")
					}

					if len(msg) == 2 && msg[0] == "del" {
						// stop tracking subscription
						log.Print("DEL MSG")
					}

					if len(msg) > 2 && len(msg) < 4 && msg[0] == "eye" {
						// create/update subscription
						log.Print("EYE MSG")
					}

					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		}
	})

	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
}
