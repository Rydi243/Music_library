package utils

import (
	"Music_library/structSong"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

// Запрос к внешнему API
func SearchSong(group, song string) (*structSong.SongDetail, error) {
	apiURL := os.Getenv("API_URL")

	encodGroup := url.QueryEscape(group) //Формируем для url без пробелов
	encodSong := url.QueryEscape(song)   //Формируем для url без пробелов

	url := fmt.Sprintf("%s?group=%s&song=%s", apiURL, encodGroup, encodSong)
	log.Printf("Отправка запроса к API:%s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Ошибка при выполнении запроса к API:%v", err)
		return nil, fmt.Errorf("Ошибка при выполнении запроса к API:%v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("API вернул статус:%s", resp.Status)
		return nil, fmt.Errorf("API вернул статус:%s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка при чтении ответа API:%v", err)
		return nil, fmt.Errorf("Ошибка при чтении ответа API:%v", err)
	}

	var detail structSong.SongDetail
	if err := json.Unmarshal(body, &detail); err != nil {
		log.Printf("Ошибка при декодировании в searchSong:%v", err)
		return nil, fmt.Errorf("Ошибка при декодировании JSON:%v", err)
	}

	return &detail, nil
}
