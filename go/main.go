package main

import (
	"fmt"
	"github.com/alexdin/firagate-teleg-bot/go/internal/listener"
	"github.com/alexdin/firagate-teleg-bot/go/internal/notify"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	_ "image/jpeg"
	"log"
	"os"
	"sync"
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

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	mqttClient := getMQTTClient()
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	bot := notify.Boot()
	go sub(mqttClient, bot)
	notify.HandleUpdates(bot)
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
	listener.Boot()
	var wg sync.WaitGroup
	wg.Add(1)
	topic := os.Getenv("MQTT_TOPIC_PREFIX")
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic %s", topic)

	if token := client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		//fmt.Printf("Subscribed to topic %s", msg.Payload())
		if eventId, ok := listener.EventHandle(msg.Payload()); ok {
			notify.SendAlarm(bot, eventId)
		}

	}); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	}

	wg.Wait()
}
