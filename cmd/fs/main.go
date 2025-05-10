package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

func downloadFileHandler(w http.ResponseWriter, r *http.Request) {
	// Указываем путь к файлу на сервере
	filePath := "./example-file.txt" // Заменить на путь к своему файлу

	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// fileName := "fname.txt"
	// Устанавливаем заголовок, чтобы указать браузеру, что файл нужно скачивать
	// w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	// w.Header().Set("Content-Type", "application/octet-stream")
	// w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Отдаем файл с помощью io.Copy
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Error sending file", http.StatusInternalServerError)
		return
	}
}

func main() {
	// Регистрируем обработчик для запроса
	http.HandleFunc("/download", downloadFileHandler)

	// Запускаем сервер на порту 8080
	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
