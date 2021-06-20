FROM golang:1.15.1


WORKDIR ./src/firagate-teleg-bot/go

RUN git clone https://github.com/alexdin/firagate-teleg-bot
COPY . .
RUN ls -la

RUN go get .