package main

import (
	"Music_library/internal/app"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Загружаем данные из .env
func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}
}

// @title Music_library
// @version 1.0
// @description Добавляет, запрашивает песни с пагинацией. запрашивает песни с пагинацией по куплетам. Обновляет и удаляет песни.
// @host localhost:8080
// @BasePath /
// @schemes http
func main() {
	app.RunDB() //Запуск контейнера postgres, ожидание 5 секунд, запуск миграций

	var err error
	app.Database, err = sql.Open("postgres", app.GetDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer app.Database.Close()

	r := mux.NewRouter()
	r.HandleFunc("/song", app.AddSong).Methods("POST")
	r.HandleFunc("/song", app.GetSong).Methods("GET")
	r.HandleFunc("/song/text", app.GetText).Methods("POST")
	r.HandleFunc("/song", app.UpdateSong).Methods("PUT")
	r.HandleFunc("/song", app.DelSong).Methods("DELETE")

	servPort := os.Getenv("SERVER_PORT")
	fmt.Printf("Server running at port:%s\n", servPort)

	errlist := http.ListenAndServe(":"+servPort, r)
	if errlist != nil {
		log.Fatal(err)
	}
}
