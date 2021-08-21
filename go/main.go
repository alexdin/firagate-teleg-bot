package main

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	_ "image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

type CamEvent struct {
	Before struct {
		ID            string        `json:"id"`
		Camera        string        `json:"camera"`
		FrameTime     float64       `json:"frame_time"`
		Label         string        `json:"label"`
		TopScore      float64       `json:"top_score"`
		FalsePositive bool          `json:"false_positive"`
		StartTime     float64       `json:"start_time"`
		EndTime       interface{}   `json:"end_time"`
		Score         float64       `json:"score"`
		Box           []int         `json:"box"`
		Area          int           `json:"area"`
		Region        []int         `json:"region"`
		CurrentZones  []interface{} `json:"current_zones"`
		EnteredZones  []interface{} `json:"entered_zones"`
		Thumbnail     interface{}   `json:"thumbnail"`
	} `json:"before"`
	After struct {
		ID            string        `json:"id"`
		Camera        string        `json:"camera"`
		FrameTime     float64       `json:"frame_time"`
		Label         string        `json:"label"`
		TopScore      float64       `json:"top_score"`
		FalsePositive bool          `json:"false_positive"`
		StartTime     float64       `json:"start_time"`
		EndTime       interface{}   `json:"end_time"`
		Score         float64       `json:"score"`
		Box           []int         `json:"box"`
		Area          int           `json:"area"`
		Region        []int         `json:"region"`
		CurrentZones  []interface{} `json:"current_zones"`
		EnteredZones  []interface{} `json:"entered_zones"`
		Thumbnail     interface{}   `json:"thumbnail"`
	} `json:"after"`
	Type string `json:"type"`
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	botKey := os.Getenv("TELEGRAM_API_KEY")

	bot, err := tgbotapi.NewBotAPI(botKey)
	if err != nil {
		log.Fatal(err)
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

	sub(mqttClient, bot)

	for update := range botUpdates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Println(update)

		//msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		//msg.ReplyToMessageID = update.Message.MessageID
		//
		//bot.Send(msg)
	}

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
		cameraEventHandler(msg.Payload(), bot)
	}); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	}

	wg.Wait()
}

func cameraEventHandler(data []byte, api *tgbotapi.BotAPI) {
	camName := os.Getenv("CAM_NAME")
	var event CamEvent
	if err := json.Unmarshal(data, &event); err != nil {
		panic(err)
	}
	if event.After.Camera == camName && event.After.Label == "person" && event.Type == "new" {
		fmt.Println("\n Detection " + event.After.ID)

		sendAlarm(api, event.After.ID)
	}
}

func sendAlarm(bot *tgbotapi.BotAPI, id string) {
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
	_, err = bot.Send(photo)
	if err != nil {
		fmt.Println(err)
	}
}
