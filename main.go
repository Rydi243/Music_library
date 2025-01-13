package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var database *sql.DB

type Song struct {
	Group       string `json:"group"`
	Song        string `json:"song"`
	Text        string `json:"text,omitempty"`
	ReleaseDate string `json:"releaseDate,omitempty"`
	Link        string `json:"link,omitempty"`
}

type SongDetail struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

func getDSN() string {
	res := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s", os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"))
	return res
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}
}

func runDB() {
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

func searchSong(group, song string) (*SongDetail, error) {
	apiURL := os.Getenv("API_URL")

	encodGroup := url.QueryEscape(group)
	encodSong := url.QueryEscape(song)

	url := fmt.Sprintf("%s?group=%s&song=%s", apiURL, encodGroup, encodSong)
	log.Printf("Отправка запроса к API: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Ошибка при выполнении запроса к API: %v", err)
		return nil, fmt.Errorf("Ошибка при выполнении запроса к API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("API вернул статус: %s", resp.Status)
		return nil, fmt.Errorf("API вернул статус: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка при чтении ответа API: %v", err)
		return nil, fmt.Errorf("Ошибка при чтении ответа API: %v", err)
	}

	var detail SongDetail
	if err := json.Unmarshal(body, &detail); err != nil {
		log.Printf("Ошибка при декодировании в searchSong: %v", err)
		return nil, fmt.Errorf("Ошибка при декодировании JSON: %v", err)
	}

	return &detail, nil
}

func addSong(w http.ResponseWriter, r *http.Request) {
	var s Song

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&s)
	if err != nil {
		http.Error(w, "Ошибка декодирования в addSong", http.StatusBadRequest)
		log.Printf("Ошибка декодирования в addSong: %v", err)
		return
	}

	var exist string
	errex := database.QueryRow(
		`SELECT song FROM public.songs WHERE "group" = $1 AND song = $2`,
		s.Group, s.Song,
	).Scan(&exist)
	if errex == nil {
		http.Error(w, "Песня уже существует", http.StatusConflict)
		return
	} else if errex != sql.ErrNoRows {
		http.Error(w, "Ошибка при проверке существования песни", http.StatusInternalServerError)
		log.Printf("Ошибка при проверке существования песни: %v", err)
		return
	}

	detail, err := searchSong(s.Group, s.Song)
	if err != nil {
		http.Error(w, "Ошибка при получении данных о песне", http.StatusInternalServerError)
		log.Printf("Ошибка при получении данных о песне: %v", err)
		return
	}

	_, errexec := database.Exec(
		`INSERT INTO public.songs ("group", song, "text", release_date, link) VALUES ($1, $2, $3, $4, $5)`,
		s.Group, s.Song, detail.Text, detail.ReleaseDate, detail.Link,
	)
	if errexec != nil {
		http.Error(w, "Ошибка вставки в базу данных", http.StatusInternalServerError)
		log.Printf("Ошибка вставки в базу данных: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode("Песня добавлена")
}

func getSong(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Group  string `json:"group"`
		Song   string `json:"song"`
		Limit  int    `json:"limit"`
		Offset int    `json:"offset"`
	}

	var filters []string
	var params []interface{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, "Ошибка декодирования в getSong", http.StatusBadRequest)
		log.Printf("Ошибка декодирования в getSong: %v", err)
		return
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	if req.Offset == 0 {
		req.Offset = 0
	}

	if req.Group != "" {
		filters = append(filters, "\"group\" = $1")
		params = append(params, req.Group)
	}
	if req.Song != "" {
		filters = append(filters, "song = $2")
		params = append(params, req.Song)
	}

	filterQuery := ""
	if len(filters) > 0 {
		query := strings.Join(filters, " AND ")
		filterQuery = "WHERE " + query
	}

	params = append(params, req.Limit, req.Offset)
	queryStr := fmt.Sprintf("SELECT \"group\", song FROM public.songs %s LIMIT $%d OFFSET $%d", filterQuery, len(params)-1, len(params))

	rows, err := database.Query(queryStr, params...)
	if err != nil {
		http.Error(w, "Ошибка запроса к базе данных", http.StatusInternalServerError)
		log.Printf("Ошибка: %v", err)
		return
	}
	defer rows.Close()

	var songs []Song
	for rows.Next() {
		var s Song
		if err := rows.Scan(&s.Group, &s.Song); err != nil {
			http.Error(w, "Ошибка обработки данных", http.StatusInternalServerError)
			log.Printf("Ошибка обработки данных: %v", err)
			return
		}
		songs = append(songs, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(songs)
}

func getText(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Song   string `json:"song"`
		Limit  int    `json:"limit"`
		Offset int    `json:"offset"`
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, "Ошибка декодирования в getText", http.StatusBadRequest)
		log.Printf("Ошибка декодирования в getText: %v", err)
		return
	}

	if req.Song == "" {
		http.Error(w, "Название песни обязательно", http.StatusBadRequest)
		return
	}

	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	var tex string
	errtext := database.QueryRow(`SELECT "text" FROM public.songs WHERE song = $1`, req.Song).Scan(&tex)
	if errtext != nil {
		if errtext == sql.ErrNoRows {
			http.Error(w, "Песня не найдена", http.StatusNotFound)
		} else {
			http.Error(w, "Ошибка при выполнении запроса к базе данных", http.StatusInternalServerError)
			log.Printf("Ошибка при выполнении запроса к базе данных в getText: %v", errtext)
		}
		return
	}

	// Заменяем литеральные \n на реальные символы новой строки
	tex = strings.ReplaceAll(tex, `\n`, "\n")

	// Разделяем текст на куплеты
	cuplet := strings.Split(tex, "\n\n")
	start := req.Offset
	end := req.Offset + req.Limit

	if start > len(cuplet) {
		start = len(cuplet)
	}
	if end > len(cuplet) {
		end = len(cuplet)
	}

	paginCuplet := cuplet[start:end]
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paginCuplet)
}

func delSong(w http.ResponseWriter, r *http.Request) {
	var s Song

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&s)
	if err != nil {
		http.Error(w, "Ошибка декодирования в delSong", http.StatusBadRequest)
		log.Println("Ошибка декодирования в delSong")
		return
	}

	if s.Song == "" {
		http.Error(w, "Название песни обязательно", http.StatusBadRequest)
		return
	}

	result, errdel := database.Exec(`DELETE FROM public.songs WHERE song = $1`, s.Song)
	if errdel != nil {
		http.Error(w, "Ошибка при удалении записи из базы данных", http.StatusInternalServerError)
		log.Printf("Ошибка при удалении записи из базы данных: %v", err)
		return
	}

	rows, errrows := result.RowsAffected()
	if errrows != nil {
		http.Error(w, "Ошибка при поиске песни для удаления", http.StatusNotFound)
		log.Printf("Ошибка при поиске песни для удаления: %v", errrows)
		return
	}
	if rows == 0 {
		http.Error(w, "Песня не найдена для удаления", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Песня удалена")
}

func updateSong(w http.ResponseWriter, r *http.Request) {
	var s Song

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&s)
	if err != nil {
		http.Error(w, "Ошибка декодирования в updateSong", http.StatusBadRequest)
		log.Println("Ошибка декодирования в updateSong")
		return
	}

	if s.Song == "" {
		http.Error(w, "Название песни обязательно", http.StatusBadRequest)
		return
	}

	result, errupd := database.Exec(`UPDATE public.songs
		 SET "group" = $1, "text" = $2 
		 WHERE song = $3`,
		s.Group, s.Text, s.Song)
	if errupd != nil {
		http.Error(w, "Ошибка обновления записи в базе данных", http.StatusInternalServerError)
		log.Printf("Ошибка обновления записи в базе данных: %v", err)
		return
	}

	rows, errrows := result.RowsAffected()
	if errrows != nil {
		http.Error(w, "Ошибка при поиске песни для обновления", http.StatusNotFound)
		log.Printf("Ошибка при поиске песни для обновления: %v", errrows)
		return
	}
	if rows == 0 {
		http.Error(w, "Песня не найдена для обновления", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Песня обновлена")
}

func main() {
	runDB()

	var err error
	database, err = sql.Open("postgres", getDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	r := mux.NewRouter()
	r.HandleFunc("/song", addSong).Methods("POST")
	r.HandleFunc("/song", getSong).Methods("GET")
	r.HandleFunc("/song/text", getText).Methods("POST")
	r.HandleFunc("/song", updateSong).Methods("PUT")
	r.HandleFunc("/song", delSong).Methods("DELETE")

	servPort := os.Getenv("SERVER_PORT")
	fmt.Printf("Server running at port:%s\n", servPort)

	errlist := http.ListenAndServe(":"+servPort, r)
	if errlist != nil {
		log.Fatal(err)
	}
}
