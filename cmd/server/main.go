package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"faustind/aubipo/pkg/models"
	"faustind/aubipo/pkg/utils"

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

	http.HandleFunc("/check-due-dates", func(w http.ResponseWriter, req *http.Request) {
		// Get all users
		// for each user get all subscriptions
		// for each subscription
		// get the one due next week
		log.Print("CHECKING DUE DATES...")
		users, err := models.GetAllUsers()
		if err != nil {
			log.Printf("AUBIPO: CHECK-DUE-DATE ERROR \n %s", err)
		}

		today := time.Now()
		for _, user := range users {
			user.Subscriptions, err = models.GetAllSubscriptionsByUser(user.ID)
			if err != nil {
				log.Printf("AUBIPO: CHECK-DUE-DATE ERROR \n %s", err)
			}
			for _, sub := range user.Subscriptions {

				lastPayDay := time.Date(sub.LastPayMonth/100, time.Month(sub.LastPayMonth%100), sub.DueDay,
					0, 0, 0, 0, time.Local)

				if today.Day() == 1 {
					if today.Month() == lastPayDay.Month() {
						msg := fmt.Sprintf("This is the last month you are planning to pay for %s.\nI will remind you to unsubscribe before the due date next month.", sub.Name)
						if _, err := bot.PushMessage(user.ID,
							linebot.NewTextMessage(msg)).Do(); err != nil {
							log.Print(err)
						}
					}
				}

				var msg string

				if utils.IsTomorrow(sub.DueDay) {
					log.Print("Due day tomorrow")
					msg = fmt.Sprintf("Your subscription to %s is due tomorrow.", sub.Name)
				} else if utils.IsInOneWeek(sub.DueDay) {
					log.Print("Due day in one week")
					msg = fmt.Sprintf("Your subscription to %s is due next week.", sub.Name)
				}

				if time.Month((sub.LastPayMonth%100)+1) == today.Month() {
					// last month was the last month
					// the user does not wish to pay the subscription for this month
					msg += fmt.Sprintf("\nYou did not plan to pay for %s this month. Don't forget to unsubscribe!", sub.Name)
				}

				if _, err := bot.PushMessage(user.ID,
					linebot.NewTextMessage(msg)).Do(); err != nil {
					log.Print(err)
				}
				continue
			}
		}

		log.Print("DONE CHECKING DUE DATES.")
		w.WriteHeader(200)
	})

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
						replyMsg := fmt.Sprintf("n subscription %d\n", len(subscriptions))
						for key, sub := range subscriptions {
							// TODO: Write this in a better format.
							msg := fmt.Sprintf("%d %s %d %d %d\n",
								key+1, sub.Name, sub.Cost, sub.DueDay, sub.LastPayMonth)
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

						var name string
						var cost, dueDay, lastMonth int

						name = msg[1]

						lastMonth, err = strconv.Atoi(msg[4])
						if err != nil || len(msg[4]) != 6 {
							log.Printf("AUBIPO: BAD LAST_MONTH VALUE: %s", msg[4])
							replyMsg := fmt.Sprintf("BAD LAST MONTH VALUE: %s", msg[4])
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMsg)).Do(); err != nil {
								log.Printf("AUBIPO: %s", err)
							}
							return
						}
						cost, err = strconv.Atoi(msg[2])
						if err != nil {
							log.Printf("AUBIPO: BAD YEN VALUE: %s", msg[2])
							replyMsg := fmt.Sprintf("BAD YEN VALUE: %s", msg[2])
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMsg)).Do(); err != nil {
								log.Printf("AUBIPO: %s", err)
							}
							return
						}

						dueDay, err = strconv.Atoi(msg[3])
						if err != nil || 31 < dueDay || dueDay < 0 {
							log.Printf("AUBIPO: BAD DUE DATE: %s", msg[3])
							replyMsg := fmt.Sprintf("BAD YEN VALUE: %s", msg[3])
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMsg)).Do(); err != nil {
								log.Printf("AUBIPO: %s", err)
							}
							return
						}

						if msg[0] == "eye" {
							// create
							CreateSub := &models.Subscription{
								Name:         name,
								UserID:       userId,
								Cost:         cost,
								DueDay:       dueDay,
								LastPayMonth: lastMonth,
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
								Name:         name,
								UserID:       userId,
								Cost:         cost,
								DueDay:       dueDay,
								LastPayMonth: lastMonth,
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
