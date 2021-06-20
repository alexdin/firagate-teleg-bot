FROM golang:1.15.1

COPY ./main.go /go/src/github.com/alexdin/firagate-teleg-bot/main.go
COPY ./.env /go/src/github.com/alexdin/firagate-teleg-bot/.env
WORKDIR /go/src/github.com/alexdin/firagate-teleg-bot
RUN go get .