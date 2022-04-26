package app

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/v7/linebot"

	"faustind/aubipo/pkg/models"
	"faustind/aubipo/pkg/utils"
)

// KitchenSink app
type KitchenSink struct {
	bot         *linebot.Client
	appBaseURL  string
	downloadDir string
}

// NewKitchenSink function
func NewKitchenSink(channelSecret, channelToken, appBaseURL string) (*KitchenSink, error) {
	apiEndpointBase := os.Getenv("ENDPOINT_BASE")
	if apiEndpointBase == "" {
		apiEndpointBase = linebot.APIEndpointBase
	}
	bot, err := linebot.New(
		channelSecret,
		channelToken,
		linebot.WithEndpointBase(apiEndpointBase), // Usually you omit this.
	)
	if err != nil {
		return nil, err
	}
	downloadDir := filepath.Join(filepath.Dir(os.Args[0]), "line-bot")
	_, err = os.Stat(downloadDir)
	if err != nil {
		if err := os.Mkdir(downloadDir, 0777); err != nil {
			return nil, err
		}
	}
	return &KitchenSink{
		bot:         bot,
		appBaseURL:  appBaseURL,
		downloadDir: downloadDir,
	}, nil
}

// Callback function for http server
func (app *KitchenSink) Callback(w http.ResponseWriter, r *http.Request) {
	events, err := app.bot.ParseRequest(r)
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
		// Send 200 response.
		w.WriteHeader(200)
		return
	}

	for _, event := range events {
		log.Printf("Got event %v", event)
		if event.Source.Type != linebot.EventSourceTypeUser {
			log.Print("AUBIPO: Event from non-user.")
			replyMsg := "Sorry, I can only talk to one user at a time!"
			if err := app.replyText(event.ReplyToken, replyMsg); err != nil {
				log.Printf("AUBIPO: %s", err)
			}
			continue
		}

		userId := event.Source.UserID

		switch event.Type {
		case linebot.EventTypeMessage:
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if err := app.handleText(message, event.ReplyToken, event.Source); err != nil {
					log.Print(err)
				}
			default:
				log.Printf("Unsupported message type: %v", event)
				if err = app.replyText(event.ReplyToken,
					"Sorry, But I don't understand!"); err != nil {
					log.Print(err)
				}
			}
		case linebot.EventTypeFollow:
			if err := app.replyText(event.ReplyToken, "Got followed event"); err != nil {
				log.Print(err)
			}
			// save user to db
			var CreateUser = &models.User{}

			CreateUser.ID = userId
			user, err := CreateUser.CreateUser()
			if err != nil {
				log.Printf("AUBIPO:CREATE_USER_ERR: %s", err)
				replyMessage := "Something wrong has happened"
				if err := app.replyText(event.ReplyToken, replyMessage); err != nil {
					log.Print(err)
				}
				return
			}

			log.Printf("AUBIPO:LINE EVENT: follow FROM %s", user.ID)

			// send instruction to set budget
			replyMessage := "You can now set your monthly budget by sending: yen AMOUNT"
			if err := app.replyText(event.ReplyToken, replyMessage); err != nil {
				log.Printf("AUBIPO:REPLY_ERR: %s", err)
			}

		case linebot.EventTypeUnfollow:
			// remove user and her subs from db
			log.Printf("Unfollowed this bot: %v", event)
			_, err = models.DeleteUser(userId)
			if err != nil {
				log.Printf("AUBIPO:UNFOLLOW ERR: \n %s", err)
			}
		default:
			log.Printf("Unknown event: %v", event)
		}
	}
}

func (app *KitchenSink) handleText(message *linebot.TextMessage, replyToken string, source *linebot.EventSource) error {
	userId := source.UserID
	if message.Text == "?" {
		// send budget and list subscriptions

		user, err := models.GetUserById(userId)
		if err != nil {
			return err
		}

		subscriptions, err := models.GetAllSubscriptionsByUser(user.ID)
		if err != nil {
			return err
		}
		replyMsg := fmt.Sprintf("n subscription %d\n", len(subscriptions))
		for key, sub := range subscriptions {
			// TODO: Write this in a better format.
			msg := fmt.Sprintf("%d %s %d %d %d\n",
				key+1, sub.Name, sub.Cost, sub.DueDay, sub.LastPayMonth)
			replyMsg += msg
		}
		if err = app.replyText(replyToken, replyMsg); err != nil {
			log.Printf("AUBIPO: %s", err)
			return err
		}
		return nil
	}

	msg := strings.Fields(strings.ToLower(message.Text))

	switch {
	case len(msg) == 2 && msg[0] == "del":
		// stop tracking subscription
		log.Printf("AUBIPO:DEL SUBSCRIPTION %s FROM %s", msg[1], userId)

		_, err := models.DeleteSubscription(userId, msg[1])
		if err != nil {
			return err
		}

		replyMsg := fmt.Sprintf("Successfully stopped tracking your subscription to %s", msg[1])

		return app.replyText(replyToken, replyMsg)

	case len(msg) == 5:
		// create/update subscription
		log.Printf("AUBIPO:EYE %s FROM %s", msg[1], userId)

		var name string
		var cost, dueDay, lastMonth int

		name = msg[1]

		lastMonth, err := strconv.Atoi(msg[4])
		if err != nil || len(msg[4]) != 6 {
			log.Printf("AUBIPO: BAD LAST_MONTH VALUE: %s", msg[4])
			replyMsg := fmt.Sprintf("BAD LAST MONTH VALUE: %s", msg[4])
			return app.replyText(replyToken, replyMsg)
		}
		cost, err = strconv.Atoi(msg[2])
		if err != nil {
			log.Printf("AUBIPO: BAD YEN VALUE: %s", msg[2])
			replyMsg := fmt.Sprintf("BAD YEN VALUE: %s", msg[2])
			return app.replyText(replyToken, replyMsg)
		}

		dueDay, err = strconv.Atoi(msg[3])
		if err != nil || 31 < dueDay || dueDay < 0 {
			log.Printf("AUBIPO: BAD DUE DATE: %s", msg[3])
			replyMsg := fmt.Sprintf("BAD YEN VALUE: %s", msg[3])
			return app.replyText(replyToken, replyMsg)
		}

		switch msg[0] {
		case "eye":
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
				log.Printf("create subscription error: %s", err)
				replyMsg := "I had a problem tracking your subscription, please try again later"
				return app.replyText(replyToken, replyMsg)
			}
			replyMsg := fmt.Sprintf("Tracking subscription to %s", name)
			return app.replyText(replyToken, replyMsg)
		case "upd":
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
				return app.replyText(replyToken, replyMsg)
			}
			replyMsg := fmt.Sprintf("Updated subscription to %s", name)
			return app.replyText(replyToken, replyMsg)
		}
	}

	return nil
}

func (app *KitchenSink) CheckDueDates(w http.ResponseWriter, r *http.Request) {
	// Get all users
	// for each user get all subscriptions
	// for each subscription
	// get the one due in one week
	// get the one due next day
	// check if this month is the last month the user wishes to pay
	// check if last month was the last month the user wished to pay
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
					if err := app.pushMessage(user.ID, msg); err != nil {
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

			if err := app.pushMessage(user.ID, msg); err != nil {
				log.Print(err)
			}
			continue
		}
	}

	w.WriteHeader(200)
}

func (app *KitchenSink) replyText(replyToken, text string) error {
	if _, err := app.bot.ReplyMessage(
		replyToken,
		linebot.NewTextMessage(text),
	).Do(); err != nil {
		return err
	}
	return nil
}

func (app *KitchenSink) pushMessage(userId, text string) error {
	if _, err := app.bot.PushMessage(
		userId,
		linebot.NewTextMessage(text)).Do(); err != nil {
		return err
	}
	return nil
}
