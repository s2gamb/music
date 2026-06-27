package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// Упрощенные структуры для декодирования Update от Telegram
type Update struct {
	UpdateID int     `json:"update_id"`
	Message  *Message `json:"message"`
}

type Message struct {
	Chat struct {
		ID int64 `json:"id"`
	} `json:"chat"`
	Text string `json:"text"`
}

func main() {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN env variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Хэндлер для вебхука. Маршрут делаем секретным (включаем токен)
	http.HandleFunc("/webhook/"+token, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var update Update
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			log.Printf("Error decoding update: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if update.Message != nil && update.Message.Text != "" {
			log.Printf("Received message: %s", update.Message.Text)
			// Обработка Deep Linking (t.me/bot?start=xyz)
			// При старте по ссылке текст будет: "/start xyz"
			go handleMessage(token, update.Message.Chat.ID, update.Message.Text)
		}

		w.WriteHeader(http.StatusOK)
	})

	log.Printf("Starting bot server on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleMessage(token string, chatID int64, text string) {
	// Простейшая отправка ответа (sendMessage)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload, _ := json.Marshal(map[string]interface{}{
		"chat_id": chatID,
		"text":    "Эхо: " + text,
	})

	_, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
