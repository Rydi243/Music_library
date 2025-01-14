package app

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	_ "github.com/lib/pq"
)

var Database *sql.DB

// Формируем DSN для бд
func GetDSN() string {
	res := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s", os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"))
	log.Printf("Сформирован DSN:%s", res)
	return res
}

// Запуск контейнера postgres, ожидание 5 секунд, запуск миграции
func RunDB() {
	errdb := exec.Command("make", "db-up").Run()
	if errdb != nil {
		log.Fatalf("Ошибка при запуске db-up postgres:%v", errdb)
	}
	log.Println("База данных запущена")

	time.Sleep(5 * time.Second)

	errmig := exec.Command("make", "migrate-up").Run()
	if errmig != nil {
		log.Fatalf("Ошибка при запуске migrate-up:%v", errmig)
	}
	log.Println("Миграция запущена")
}
