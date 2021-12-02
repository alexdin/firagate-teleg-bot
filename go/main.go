package main

import (
	"./internal/events"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	_ "image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔇", "muteCommand"),
		tgbotapi.NewInlineKeyboardButtonData("🔉", "unMuteCommand"),
	),
)

var muteTime int64 = time.Now().Unix()

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

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

	mqttClient := getMQTTClient()

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	go sub(mqttClient, bot)

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

}

func unMute() {
	muteTime = time.Now().Unix() - 1
}

func getMQTTClient() mqtt.Client {

	broker := os.Getenv("MQTT_BROKER")
	port := os.Getenv("MQTT_PORT")
	fmt.Println("Try connect:", broker, port)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%s", broker, port))
	opts.SetClientID(os.Getenv("MQTT_CLIENT"))

	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	return mqtt.NewClient(opts)
}

func sub(client mqtt.Client, bot *tgbotapi.BotAPI) {
	var wg sync.WaitGroup
	wg.Add(1)
	topic := os.Getenv("MQTT_TOPIC_PREFIX")
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic %s", topic)

	if token := client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		//fmt.Printf("Subscribed to topic %s", msg.Payload())
		if eventId, ok := events.EventHandle(msg.Payload()); ok {
			sendAlarm(bot, eventId)
		}

	}); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	}

	wg.Wait()
}

func sendAlarm(bot *tgbotapi.BotAPI, id string) {
	if muteTime < time.Now().Unix() {
		sendPhoto(bot, id)
	}
}

func sendPhoto(bot *tgbotapi.BotAPI, id string) {
	channelId := os.Getenv("TELEGRAM_CHANNEL_ID")
	channelIdInt, _ := strconv.ParseInt(channelId, 10, 64)
	frigateUrl := os.Getenv("FRIGATE_URL")

	fullPath := frigateUrl + "/api/events/" + id + "/snapshot.jpg"

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

func mute() {
	muteTime = time.Now().Unix() + 300
}
