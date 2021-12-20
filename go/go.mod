module firagate-teleg-bot

go 1.15

replace (
	internal/listener => ./internal/listener
	internal/notify => ./internal/notify
)

require (
	github.com/eclipse/paho.mqtt.golang v1.3.5 // indirect
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible // indirect
	github.com/joho/godotenv v1.4.0 // indirect
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	internal/listener v0.0.0-00010101000000-000000000000 // indirect
)
