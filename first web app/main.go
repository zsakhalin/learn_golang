package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var tmpl = template.Must(template.ParseFiles("index.html")) // переменная уровня пакета, которая указывает на определение шаблона из предоставленных файлов
// template.ParseFiles анализирует файл index.html в корне каталога проекта и проверяет его на валидность
var apiKey *string // переменная для передачи токена API в виде флага командной строки

// модель данных, получаемых от News API
// https://newsapi.org/docs/endpoints/everything
type Source struct {
	ID   interface{} `json:"id"`
	Name string      `json:"name"`
}
type Article struct {
	Source      Source    `json:"source"`
	Author      string    `json:"author"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	URLToImage  string    `json:"urlToImage"`
	PublishedAt time.Time `json:"publishedAt"`
	Content     string    `json:"content"`
}
type Results struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}
type Search struct {
	SearchKey  string  // поисковый запрос
	NextPage   int     // позволяет пролистывать результаты
	TotalPages int     // общее количество страниц результатов запроса
	Results    Results // текущая страница результатов запроса
}

// функция-обработчик для корневого пути /
// w — это структура для отправки ответов на HTTP-запрос
// r - HTTP-запрос, полученный от клиента (доступ к данным, отправляемым веб-браузером на сервере)
func indexHandler(w http.ResponseWriter, r *http.Request) {
	// w.Write([]byte("<h1>HELLO W</h1>")) //принимает слайс байтов и записывает объединенные данные как часть HTTP-ответа
	tmpl.Execute(w, nil)
}

// Создаем роут /search, который обрабатывает поисковые запросы для новостных статей
// извлекает параметры q и page из URL-адреса запроса и выводит в терминал
func searchHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	params := u.Query()
	searchKey := params.Get("q") // q - запрос пользователя
	page := params.Get("page")   // page используется для пролистывания результатов
	if page == "" {              // Если он не включен в URL, присвоим 1
		page = "1"
	}

	fmt.Println("Search Query is: ", searchKey)
	fmt.Println("Resault page is: ", page)

	search := &Search{}          //  новый экземпляр структуры Searc
	search.SearchKey = searchKey // устанавливаем значение поля SearchKey равным значению параметра URL q в HTTP-запросе

	next, err := strconv.Atoi(page) //конвертируем переменную page в целое число
	if err != nil {
		http.Error(w, "Unexpected server error", http.StatusInternalServerError)
		return
	}

	search.NextPage = next //присваиваем int(page) полю NextPage переменной search
	pageSize := 20         // количество результатов, которые API новостей будет возвращать в ответ

	endpoint := fmt.Sprintf("https://newsapi.org/v2/everything?q=%s&pageSize=%d&page=%d&apiKey=%s&sortBy=publishedAt&language=en", url.QueryEscape(search.SearchKey), pageSize, search.NextPage, *apiKey)
	resp, err := http.Get(endpoint)
	if err != nil {
		w.WriteHeader((http.StatusInternalServerError))
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 { // Если ответ от News API не 200 OK
		w.WriteHeader((http.StatusInternalServerError))
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&search.Results)
	if err != nil {
		w.WriteHeader((http.StatusInternalServerError))
		return
	}

	search.TotalPages = int(math.Ceil(float64(search.Results.TotalResults / pageSize))) // общее количество страниц = TotalResults / pageSize
	err = tmpl.Execute(w, search)                                                       // рендерим шаблон и передаем переменную search в качестве интерфейса данных, это даёт доступ к данным из объекта JSON в шаблоне                                                  //
	if err != nil {
		w.WriteHeader((http.StatusInternalServerError))
		return
	}
}

func main() {
	// передача токена API в виде флага командной строки
	// go run main.go -apikey=<newsapi access key>
	apiKey = flag.String("apikey", "", "Newsapi.org access key") // определение строкового флага
	flag.Parse()
	if *apiKey == "" {
		log.Fatal("API key is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" // http://localhost:3000
	}

	mux := http.NewServeMux() //новый мультиплексор HTTP-запросов, мультиплексор запросов сопоставляет URL-адрес входящих запросов со списком зарегистрированных путей и вызывает соответствующий обработчик для пути всякий раз, когда найдено совпадение.

	fs := http.FileServer(http.Dir("assets"))                // экземпляр объекта файлового сервера, c каталогом, в котором находятся все статические файлы
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs)) // указание маршрутизатору использовать этот fs объект файлового сервера для всех путей, начинающихся с префикса /assets/
	mux.HandleFunc("/search/", searchHandler)                // регистрируем обработчик для пути /search
	mux.HandleFunc("/", indexHandler)                        // регистрируем обработчик для пути /
	http.ListenAndServe(":"+port, mux)                       //запускает сервер на порту 3000, если порт не установлен окружением
}
