package main

import (
	// "fmt"
	"html/template"
	"net/http"
	"os"
)

var tmpl = template.Must(template.ParseFiles("index.html")) // переменная уровня пакета, которая указывает на определение шаблона из предоставленных файлов
// template.ParseFiles анализирует файл index.html в корне каталога проекта и проверяет его на валидность

// функция-обработчик для корневого пути /
// w — это структура для отправки ответов на HTTP-запрос
// r - HTTP-запрос, полученный от клиента (доступ к данным, отправляемым веб-браузером на сервере)
func indexHandler(w http.ResponseWriter, r *http.Request) {
	// w.Write([]byte("<h1>HELLO W</h1>")) //принимает слайс байтов и записывает объединенные данные как часть HTTP-ответа
	tmpl.Execute(w, nil)
	// fmt.Println(w)
}

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" // http://localhost:3000
	}

	mux := http.NewServeMux() //новый мультиплексор HTTP-запросов, мультиплексор запросов сопоставляет URL-адрес входящих запросов со списком зарегистрированных путей и вызывает соответствующий обработчик для пути всякий раз, когда найдено совпадение.

	fs := http.FileServer(http.Dir("assets"))                // экземпляр объекта файлового сервера, c каталогом, в котором находятся все статические файлы
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs)) // указание маршрутизатору использовать этот fs объект файлового сервера для всех путей, начинающихся с префикса /assets/
	mux.HandleFunc("/", indexHandler)
	http.ListenAndServe(":"+port, mux) //запускает сервер на порту 3000, если порт не установлен окружением
}
