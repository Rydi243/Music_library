package app

import (
	"Music_library/internal/utils"
	"Music_library/structSong"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "github.com/lib/pq"
)

// @Summary Добавить песню(обращается к тестовому апи)
// @Description Добавляет новую песню с тестового апи
// @Accept json
// @Produce json
// @Param song body structSong.Song true "Данные песни"
// @Success 201 {string} string "Песня добавлена"
// @Failure 400 {string} string "Ошибка декодирования"
// @Failure 409 {string} string "Песня уже существует"
// @Failure 500 {string} string "Ошибка при получении данных о песне или вставке в базу данных"
// @Router /song [post]
func AddSong(w http.ResponseWriter, r *http.Request) {
	var s structSong.Song

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&s)
	if err != nil {
		http.Error(w, "Ошибка декодирования в addSong", http.StatusBadRequest)
		log.Printf("Ошибка декодирования в addSong:%v", err)
		return
	}

	var exist string
	errex := Database.QueryRow(
		`SELECT song FROM public.songs WHERE "group" = $1 AND song = $2`,
		s.Group, s.Song,
	).Scan(&exist)
	if errex == nil {
		http.Error(w, "Песня уже существует", http.StatusConflict)
		return
	} else if errex != sql.ErrNoRows {
		http.Error(w, "Ошибка при проверке существования песни", http.StatusInternalServerError)
		log.Printf("Ошибка при проверке существования песни:%v", err)
		return
	}

	detail, err := utils.SearchSong(s.Group, s.Song)
	if err != nil {
		http.Error(w, "Ошибка при получении данных о песне", http.StatusInternalServerError)
		log.Printf("Ошибка при получении данных о песне:%v", err)
		return
	}

	_, errexec := Database.Exec(
		`INSERT INTO public.songs ("group", song, "text", release_date, link) VALUES ($1, $2, $3, $4, $5)`,
		s.Group, s.Song, detail.Text, detail.ReleaseDate, detail.Link,
	)
	if errexec != nil {
		http.Error(w, "Ошибка вставки в базу данных", http.StatusInternalServerError)
		log.Printf("Ошибка вставки в базу данных:%v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode("Песня добавлена")
}

// @Summary Получить список песен
// @Description Возвращает список песен с пагинацией и фильтрацией по всем полям
// @Accept json
// @Produce json
// @Param request body structSong.ReqGetSong true "Параметры"
// @Success 200 {array} structSong.Song "Список песен"
// @Failure 400 {string} string "Ошибка декодирования"
// @Failure 500 {string} string "Ошибка запроса к базе данных"
// @Router /song [get]
func GetSong(w http.ResponseWriter, r *http.Request) {
	var req structSong.ReqGetSong

	var filters []string
	var params []interface{}
	paramIndex := 1 // Счетчик для индексов параметров

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, "Ошибка декодирования в getSong", http.StatusBadRequest)
		log.Printf("Ошибка декодирования в getSong:%v", err)
		return
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	if req.Group != "" {
		filters = append(filters, fmt.Sprintf("\"group\" = $%d", paramIndex))
		params = append(params, req.Group)
		paramIndex++
	}
	if req.Song != "" {
		filters = append(filters, fmt.Sprintf("song = $%d", paramIndex))
		params = append(params, req.Song)
		paramIndex++
	}
	if req.Text != "" {
		filters = append(filters, fmt.Sprintf("text ILIKE $%d", paramIndex))
		params = append(params, "%"+req.Text+"%") // Поиск по частичному совпадению
		paramIndex++
	}
	if req.ReleaseDate != "" {
		filters = append(filters, fmt.Sprintf("release_date = $%d", paramIndex))
		params = append(params, req.ReleaseDate)
		paramIndex++
	}
	if req.Link != "" {
		filters = append(filters, fmt.Sprintf("link = $%d", paramIndex))
		params = append(params, req.Link)
		paramIndex++
	}

	filterQuery := ""
	if len(filters) > 0 {
		filterQuery = "WHERE " + strings.Join(filters, " AND ")
	}

	params = append(params, req.Limit, req.Offset)
	queryStr := fmt.Sprintf(
		`SELECT "group", song, text, release_date, link FROM public.songs %s LIMIT $%d OFFSET $%d`,
		filterQuery,
		paramIndex,
		paramIndex+1,
	)

	rows, err := Database.Query(queryStr, params...)
	if err != nil {
		http.Error(w, "Ошибка запроса к базе данных", http.StatusInternalServerError)
		log.Printf("Ошибка: %v", err)
		return
	}
	defer rows.Close()

	var songs []structSong.Song
	for rows.Next() {
		var s structSong.Song
		errrow := rows.Scan(&s.Group, &s.Song, &s.Text, &s.ReleaseDate, &s.Link)
		if errrow != nil {
			http.Error(w, "Ошибка обработки данных", http.StatusInternalServerError)
			log.Printf("Ошибка обработки данных: %v", errrow)
			return
		}
		songs = append(songs, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(songs)
}

// @Summary Получить текст песни с пагинацией по куплетам
// @Description Возвращает текст песни, разделенный на куплеты, с пагинацией
// @Accept json
// @Produce json
// @Param request body structSong.ReqTextSong true "Параметры"
// @Success 200 {array} string "Список куплетов"
// @Failure 400 {string} string "Ошибка декодирования или отсутствует название песни"
// @Failure 404 {string} string "Песня не найдена"
// @Failure 500 {string} string "Ошибка при выполнении запроса к базе данных"
// @Router /song/text [post]
func GetText(w http.ResponseWriter, r *http.Request) {
	var req structSong.ReqTextSong

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, "Ошибка декодирования в getText", http.StatusBadRequest)
		log.Printf("Ошибка декодирования в getText:%v", err)
		return
	}

	if req.Song == "" {
		http.Error(w, "Название песни обязательно", http.StatusBadRequest)
		return
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	var tex string
	errtext := Database.QueryRow(`SELECT "text" FROM public.songs WHERE song = $1`, req.Song).Scan(&tex)
	if errtext != nil {
		if errtext == sql.ErrNoRows {
			http.Error(w, "Песня не найдена", http.StatusNotFound)
		} else {
			http.Error(w, "Ошибка при выполнении запроса к базе данных", http.StatusInternalServerError)
			log.Printf("Ошибка при выполнении запроса к базе данных в getText:%v", errtext)
		}
		return
	}

	tex = strings.ReplaceAll(tex, `\n`, "\n") // Заменяем символ \n на реальный \n, т.к бд не видит при пагинации разделения
	cuplet := strings.Split(tex, "\n\n")      // Разделяем текст на куплеты
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

// @Summary Удалить песню
// @Description Удаляет песню из библиотеки по её названию
// @Accept json
// @Produce json
// @Param song body structSong.Song true "Название песни для удаления"
// @Success 200 {string} string "Песня удалена"
// @Failure 400 {string} string "Ошибка декодирования или отсутствует название песни"
// @Failure 404 {string} string "Песня не найдена"
// @Failure 500 {string} string "Ошибка при удалении записи из базы данных"
// @Router /song [delete]
func DelSong(w http.ResponseWriter, r *http.Request) {
	var s structSong.Song

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

	result, errdel := Database.Exec(`DELETE FROM public.songs WHERE song = $1`, s.Song)
	if errdel != nil {
		http.Error(w, "Ошибка при удалении записи из базы данных", http.StatusInternalServerError)
		log.Printf("Ошибка при удалении записи из базы данных:%v", err)
		return
	}

	rows, errrows := result.RowsAffected()
	if errrows != nil {
		http.Error(w, "Ошибка при поиске песни для удаления", http.StatusNotFound)
		log.Printf("Ошибка при поиске песни для удаления:%v", errrows)
		return
	}
	if rows == 0 {
		http.Error(w, "Песня не найдена для удаления", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Песня удалена")
}

// @Summary Обновить данные песни
// @Description Обновляет информацию о песне
// @Accept json
// @Produce json
// @Param song body structSong.Song true "Информация для обновления"
// @Success 200 {string} string "Песня обновлена"
// @Failure 400 {string} string "Ошибка декодирования или отсутствуют данные для обновления"
// @Failure 404 {string} string "Песня не найдена"
// @Failure 500 {string} string "Ошибка обновления записи в базе данных"
// @Router /song [put]
func UpdateSong(w http.ResponseWriter, r *http.Request) {
	var s structSong.Song

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

	// Формируем динамический запрос для базки
	query := `UPDATE public.songs SET `
	var params []interface{}
	paramCount := 1

	if s.Group != "" {
		query += `"group" = $` + fmt.Sprint(paramCount) + `, `
		params = append(params, s.Group)
		paramCount++
	}
	if s.Text != "" {
		query += `"text" = $` + fmt.Sprint(paramCount) + `, `
		params = append(params, s.Text)
		paramCount++
	}
	if s.ReleaseDate != "" {
		query += `release_date = $` + fmt.Sprint(paramCount) + `, `
		params = append(params, s.ReleaseDate)
		paramCount++
	}
	if s.Link != "" {
		query += `link = $` + fmt.Sprint(paramCount) + `, `
		params = append(params, s.Link)
		paramCount++
	}

	if len(params) == 0 {
		http.Error(w, "Нет данных для обновления", http.StatusBadRequest)
		return
	}

	// Убираем последнюю запятую, перед WHERE
	query = query[:len(query)-2]

	query += ` WHERE song = $` + fmt.Sprint(paramCount)
	params = append(params, s.Song)

	result, errupd := Database.Exec(query, params...)
	if errupd != nil {
		http.Error(w, "Ошибка обновления записи в базе данных", http.StatusInternalServerError)
		log.Printf("Ошибка обновления записи в базе данных:%v", errupd)
		return
	}

	rows, errrows := result.RowsAffected()
	if errrows != nil {
		http.Error(w, "Ошибка при поиске песни для обновления", http.StatusNotFound)
		log.Printf("Ошибка при поиске песни для обновления:%v", errrows)
		return
	}
	if rows == 0 {
		http.Error(w, "Песня не найдена для обновления", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Песня обновлена")
}
