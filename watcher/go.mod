module watcher

go 1.15

require internal/notifyer v1.0.0

replace internal/notifyer => ./internal/notifyer

require (
	github.com/eclipse/paho.mqtt.golang v1.3.5 // indirect
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible // indirect
	github.com/joho/godotenv v1.4.0 // indirect
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	internal/listener v1.0.0
)

replace internal/listener => ./internal/listener
