package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
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
		log.Fatalf("AUBIPO: %s", err)
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
			if event.Source.Type != linebot.EventSourceTypeUser {
				log.Print("AUBIPO: Event from non-user.")
				replyMsg := "Sorry, I can only talk to one user at a time!"
				if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMsg)).Do(); err != nil {
					log.Printf("AUBIPO: %s", err)
				}
				return
			}
			userId := event.Source.UserID

			switch event.Type {
			case linebot.EventTypeFollow:
				// save user to db
				var CreateUser = &models.User{}

				CreateUser.ID = userId
				user, err := CreateUser.CreateUser()
				if err != nil {
					log.Printf("AUBIPO:CREATE_USER_ERR: %s", err)
					replyMessage := "Something wrong has happened"
					if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMessage)).Do(); err != nil {
						log.Print(err)
					}
					return
				}

				log.Printf("AUBIPO:LINE EVENT: follow FROM %s", user.ID)

				// send instruction to set budget
				replyMessage := "You can now set your monthly budget by sending: yen AMOUNT"
				if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMessage)).Do(); err != nil {
					log.Printf("AUBIPO:REPLY_ERR: %s", err)
				}

			case linebot.EventTypeUnfollow:
				// remove user and her subs from db
				log.Printf("AUBIPO:LINE EVENT: unfollow FROM %s", userId)

				_, err = models.DeleteUser(userId)
				if err != nil {
					log.Printf("AUBIPO:UNFOLLOW ERR: \n %s", err)
				}

			case linebot.EventTypeMessage:
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					if message.Text == "?" {
						// send budget and list subscriptions

						user, err := models.GetUserById(userId)
						if err != nil {
							log.Print(err)
							return
						}

						subscriptions, err := models.GetAllSubscriptionsByUser(user.ID)
						if err != nil {
							log.Print(err)
							return
						}
						replyMsg := fmt.Sprintf("budget %d, n subscription %d\n", user.Budget, len(subscriptions))
						for key, sub := range subscriptions {
							msg := fmt.Sprintf("%d %s %d %s %s\n",
								key, sub.Name, sub.Cost, sub.StartDate, sub.EndDate)
							replyMsg += msg
						}
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMsg)).Do(); err != nil {
							log.Printf("AUBIPO: %s", err)
						}
						return
					}
					msg := strings.Fields(strings.ToLower(message.Text))

					if len(msg) == 2 && msg[0] == "yen" {
						// update user budget
						log.Print("AUBIPO: YEN MSG")
						amount, err := strconv.Atoi(msg[1])
						if err != nil {
							log.Printf("AUBIPO: BAD YEN VALUE: %s", msg[1])
							replyMsg := fmt.Sprintf("BAD YEN VALUE: %s", msg[1])
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMsg)).Do(); err != nil {
								log.Printf("AUBIPO: %s", err)
							}
							return
						}

						user, err := models.GetUserById(userId)
						if err != nil {
							log.Print(err)
						}
						user.Budget = amount
						_, err = user.UpdateUser()
						if err != nil {
							log.Print(err)
						}

						replyMsg := fmt.Sprintf("Your monthly budget is set to %d ¥", amount)
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMsg)).Do(); err != nil {
							log.Printf("AUBIPO: %s", err)
						}
						return
					}

					if len(msg) == 2 && msg[0] == "del" {
						// stop tracking subscription
						log.Printf("AUBIPO:DEL SUBSCRIPTION %s FROM %s", msg[1], userId)

						_, err = models.DeleteSubscription(userId, msg[1])
						if err != nil {
							log.Print(err)
						}

						replyMsg := fmt.Sprintf("Successfully stopped tracking your subscription to %s", msg[1])

						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMsg)).Do(); err != nil {
							log.Print(err)
						}
						return
					}

					if len(msg) == 4 || len(msg) == 5 {
						// create/update subscription
						log.Printf("AUBIPO:EYE %s FROM %s", msg[1], userId)

						var name, startDate, endDate string
						var cost int

						name, startDate = msg[1], msg[3]

						cost, err = strconv.Atoi(msg[2])
						if err != nil {
							log.Printf("AUBIPO: BAD YEN VALUE: %s", msg[2])
							replyMsg := fmt.Sprintf("BAD YEN VALUE: %s", msg[2])
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMsg)).Do(); err != nil {
								log.Printf("AUBIPO: %s", err)
							}
							return
						}
						if len(msg) == 5 {
							endDate = msg[4]
						}

						if msg[0] == "eye" {
							// create
							CreateSub := &models.Subscription{
								Name:      name,
								UserID:    userId,
								Cost:      cost,
								StartDate: startDate,
								EndDate:   endDate,
							}
							_, err := CreateSub.CreateSubscription()
							if err != nil {
								log.Printf("AUBIPO:CREATE_ERR: %s", err)
								replyMsg := "problem updating subscription, pls try again later"
								if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMsg)).Do(); err != nil {
									log.Print(err)
								}
							}
							replyMsg := fmt.Sprintf("Tracking subscription to %s", name)
							if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMsg)).Do(); err != nil {
								log.Print(err)
							}

							return
						} else if msg[0] == "upd" {
							// update
							UpdateSub := &models.Subscription{
								Name:      name,
								UserID:    userId,
								Cost:      cost,
								StartDate: startDate,
								EndDate:   endDate,
							}
							_, err := UpdateSub.UpdateSubscription()
							if err != nil {
								log.Printf("AUBIPO:CREATE_ERR: %s", err)
								replyMsg := "problem updating subscription, pls try again later"
								if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMsg)).Do(); err != nil {
									log.Print(err)
								}
							}
							replyMsg := fmt.Sprintf("Updated subscription to %s", name)
							if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMsg)).Do(); err != nil {
								log.Print(err)
							}
							return
						}
					}

					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("Sorry, But I don't understand!")).Do(); err != nil {
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
