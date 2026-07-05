package main

import (
	"fmt"
	"log"
	"time"

	"amazingbot/sampquery"

	tele "gopkg.in/telebot.v3"
)
var cfg Config
func main() {
	cfg = LoadConfig()
	pref := tele.Settings{
		Token:  cfg.BotToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatalf("Не удалось создать бота: %v", err)
	}

	bot.Handle("/start", handleStart)
	bot.Handle("/status", handleStatus)
	bot.Handle("/ping", handlePing)

	log.Println("Бот запущен...")
	bot.Start()
}

func handleStart(c tele.Context) error {
	text := "Привет! Я показываю статистику SA-MP сервера.\n\n" +
		"Команды:\n" +
		"/status — онлайн и инфа о сервере\n" +
		"/ping — пинг до сервера"
	return c.Send(text)
}

func handleStatus(c tele.Context) error {
	info, err := sampquery.GetInfo(cfg.ServerIP, cfg.ServerPort)
	if err != nil {
		return c.Send(fmt.Sprintf("⚠️ Не удалось получить данные: %v", err))
	}

	passwordText := "нет"
	if info.Password {
		passwordText = "да"
	}

	text := fmt.Sprintf(
		"<b>%s</b>\n\n"+
			"Онлайн: %d/%d\n"+
			"Режим: %s\n"+
			"Язык: %s\n"+
			"Пароль: %s",
		info.Hostname, info.Players, info.MaxPlayers,
		info.Gamemode, info.Language, passwordText,
	)

	return c.Send(text, tele.ModeHTML)
}

func handlePing(c tele.Context) error {
	ping, err := sampquery.GetPing(cfg.ServerIP, cfg.ServerPort)
	if err != nil {
		return c.Send(fmt.Sprintf("⚠️ Не удалось получить данные: %v", err))
	}

	text := fmt.Sprintf("Пинг до сервера: %.1f мс", float64(ping.Microseconds())/1000.0)
	return c.Send(text)
}
