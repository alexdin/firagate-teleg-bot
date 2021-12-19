package notify

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔇", "muteCommand"),
		tgbotapi.NewInlineKeyboardButtonData("🔉", "unMuteCommand"),
	),
)

var muteTime int64 = time.Now().Unix()

func Boot() *tgbotapi.BotAPI {
	botKey := os.Getenv("TELEGRAM_API_KEY")

	bot, err := tgbotapi.NewBotAPI(botKey)
	if err != nil {
		for {
			bot, err = tgbotapi.NewBotAPI(botKey)
			if err != nil {
				switch err.(type) {
				case *url.Error:
					log.Println("Internet is dead :( retrying to connect in 2 minutes")
					time.Sleep(1 * time.Minute)
				default:
					log.Fatal(err)
				}
			} else {
				break
			}
		}
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	botUpdates, err := bot.GetUpdatesChan(u)
	// botUpdates from  bot telegram

	for update := range botUpdates {
		if update.CallbackQuery != nil {
			fmt.Print(update)

			bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
			switch update.CallbackQuery.Data {
			case "muteCommand":
				mute()
			case "unMuteCommand":
				unMute()
			}
		}
		if update.Message != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			fmt.Println(msg)
		}
	}
	return bot
}

func unMute() {
	muteTime = time.Now().Unix() - 1
}

func mute() {
	muteTime = time.Now().Unix() + 300
}

func SendAlarm(bot *tgbotapi.BotAPI, id string) {
	if muteTime < time.Now().Unix() {
		sendPhoto(bot, id)
	}
}

func sendPhoto(bot *tgbotapi.BotAPI, id string) {
	channelId := os.Getenv("TELEGRAM_CHANNEL_ID")
	channelIdInt, _ := strconv.ParseInt(channelId, 10, 64)
	frigateUrl := os.Getenv("FRIGATE_URL")

	fullPath := frigateUrl + "/api/listener/" + id + "/snapshot.jpg"

	res, err := http.Get(fullPath)

	if err != nil {
		fmt.Println(err)
	}

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}
	bytes := tgbotapi.FileBytes{Name: "image.jpg", Bytes: content}
	photo := tgbotapi.NewPhotoUpload(channelIdInt, bytes)
	photo.ReplyMarkup = numericKeyboard
	_, err = bot.Send(photo)
	if err != nil {
		fmt.Println(err)
	}
}
