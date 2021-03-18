package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type TelegramMsg struct {
	ChatId    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}
type Message struct {
	Datetime string
	Level    string
	Msg      string
}

var baseTelegramUrl string = "https://api.telegram.org/bot"

func SendMessage(msg, chatId, token string) {
	tMsg := TelegramMsg{
		chatId, msg, "Markdown",
	}
	jsonValue, _ := json.Marshal(tMsg)
	_, err := http.Post(baseTelegramUrl+token+"/sendmessage", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Println("Unable to send telegram msg", msg)
	}
}
