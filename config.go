package main

import (
	"log"
	"os"
	"strconv"
 
	"github.com/joho/godotenv"
)
type Config struct {
	BotToken string
	ServerIP string
	ServerPort uint16
}

func LoadConfig() Config{
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, читаю переменные окружения напрямую")
	}
 
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN не задан — добавь его в .env или переменные окружения")
	}
 
	ip := os.Getenv("SERVER_IP")
	if ip == "" {
		log.Fatal("SERVER_IP не задан")
	}
 
	portStr := os.Getenv("SERVER_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("SERVER_PORT некорректен: %v", err)
	}
 
	return Config{
		BotToken:   token,
		ServerIP:   ip,
		ServerPort: uint16(port),
	}
}
