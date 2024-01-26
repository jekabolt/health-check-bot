package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	mutex       sync.Mutex
	chats       = make(map[int64]bool)
	messageWait = 30 * time.Second // Initial timer duration
)

func main() {
	chats[-1002074292263] = true
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Panic("Telegram bot token not set")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 10

	go messageLoop(bot)

	// Keep the bot running
	select {}
}

func messageLoop(bot *tgbotapi.BotAPI) {
	for {
		time.Sleep(messageWait)
		if err := checkHealth(); err != nil {
			log.Println(err)
			sendMessageToAllChats(bot, "Backend is down: "+err.Error())
			// Increase timer by 15 minutes, up to a maximum of 1 hour
			if messageWait < time.Hour {
				messageWait += 15 * time.Minute
			}
		} else {
			// Reset timer to 30 seconds if no error
			messageWait = 30 * time.Second
		}
	}
}

func checkHealth() error {
	resp, err := http.Get("http://backend.grbpwr.com:8081")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}
	return nil
}

func sendMessageToAllChats(bot *tgbotapi.BotAPI, message string) {
	mutex.Lock()
	defer mutex.Unlock()

	for chatID := range chats {
		msg := tgbotapi.NewMessage(chatID, message)
		if _, err := bot.Send(msg); err != nil {
			log.Println("Failed to send message to chat", chatID, ":", err)
		}
	}
}
