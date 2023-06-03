package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"text/template"
)

var db *sql.DB
var dbOnce sync.Once
var dbErr error

// Подключение к БД
func connectionDb() (*sql.DB, error) {
	dbOnce.Do(func() {
		db, dbErr = sql.Open("mysql", "root:@/pet_projects") //newuser:password@/pet_projects
		if dbErr != nil {
			log.Printf("Ошибка подключения к базе данных: %v", dbErr)
			return
		}
	})

	return db, dbErr
}

type Data struct {
	ID            int
	EnteredValue  string
	Result        string
	RecordingDate string
}

type DataToDoList struct {
	ID        int
	Name      string
	Text      string
	DateAdded string
}

func home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	files := []string{
		"../../ui/html/home.page.html",
		"../../ui/html/baseHome.layout.html",
		"../../ui/html/footer.partial.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Внутренная ошибка на сервере", 500)
	}

	err = ts.Execute(w, nil)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Внутренняя ошибка на сервере2", 500)
		return
	}
}

// Calculator
func calculator(w http.ResponseWriter, r *http.Request) {

	files := []string{
		"../../ui/html/calculator.page.html",
		"../../ui/html/base.layout.html",
		"../../ui/html/footer.partial.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Внутренняя ошибка на сервере", 500)
		return
	}

	//Вызов функции подключение к БД
	_, err = connectionDb()
	if err != nil {
		log.Println(err)
	}

	//Вывод данных из БД на страницу
	rows, err := db.Query("SELECT id, entered_value, result, recording_date FROM calculator ORDER BY id DESC LIMIT 10")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	// Создаем срез для хранения всех объектов Data
	var dataRecords []Data

	// Итерируем по строкам и заполняем структуру данными
	for rows.Next() {
		var dataRecord Data

		err = rows.Scan(&dataRecord.ID, &dataRecord.EnteredValue, &dataRecord.Result, &dataRecord.RecordingDate)
		if err != nil {
			log.Fatal(err)
		}

		// Добавляем объект Data в срез
		dataRecords = append(dataRecords, dataRecord)
	}

	err = ts.Execute(w, dataRecords)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Внутренняя ошибка на сервере2", 500)
		return
	}
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	//Проверяем, что метод запроса является POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Используйте r.FormValue для получения значений полей формы
	data := r.FormValue("data")
	result := parseTheReceivedValue(data) //Приводим строку в числовой тип

	// Обрабатываем полученные данные (например, сохраняем в базу данных)
	//Вызов функции подключение к БД
	_, err := connectionDb()
	if err != nil {
		return
	}

	//Если веденные данные больше 10 символов обрезать
	if len(data) > 10 {
		data = data[:10]
	}

	//Добавляем данные в БД
	_, err = db.Exec("INSERT INTO calculator (entered_value, result) VALUES (?, ?)", data, result)
	if err != nil {
		log.Fatal(err)
	}
	http.Redirect(w, r, "/calculator", http.StatusSeeOther)
}

// Распарить полученное значение
func parseTheReceivedValue(dataParam string) float64 {
	var arithmeticallyParameter string
	for _, v := range dataParam {
		switch string(v) {
		case "+":
			arithmeticallyParameter = "+"
		case "-":
			arithmeticallyParameter = "-"
		case "*":
			arithmeticallyParameter = "*"
		case "/":
			arithmeticallyParameter = "/"
		case "%":
			arithmeticallyParameter = "%"
		}
	}

	strSplit := strings.Split(dataParam, arithmeticallyParameter)
	var result float64

	firstNumber, _ := strconv.ParseFloat(strSplit[0], 64)
	secondNumber, _ := strconv.ParseFloat(strSplit[1], 64)

	switch arithmeticallyParameter {
	case "+":
		result = firstNumber + secondNumber
	case "-":
		result = firstNumber - secondNumber
	case "*":
		result = firstNumber * secondNumber
	case "/":
		result = firstNumber / secondNumber
	case "%":
		result = float64(int(firstNumber) % int(secondNumber))
	}
	return result
}

func deleteEntry(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Fprintf(w, "id: %d", id)

	//Вызов функции подключение к БД
	_, err = connectionDb()
	if err != nil {
		return
	}

	_, err = db.Exec("DELETE FROM calculator WHERE id = ?", id)
	if err != nil {
		log.Fatal(err)
	}
	http.Redirect(w, r, "/calculator", http.StatusSeeOther)
}

// To do list
func toDoList(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"../../ui/html/toDoList.page.html",
		"../../ui/html/base.layout.html",
		"../../ui/html/footer.partial.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Внутренняя ошибка на сервере", 500)
		return
	}

	//Вызов функции подключение к БД
	_, err = connectionDb()
	if err != nil {
		log.Println(err)
	}

	//Вывод данных из БД на страницу
	rows, err := db.Query("SELECT id, name, text, date_added FROM todolist ORDER BY id DESC")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	// Создаем срез для хранения всех объектов Data
	var dataRecordsToDoList []DataToDoList

	// Итерируем по строкам и заполняем структуру данными
	for rows.Next() {
		var dataRecordToDoList DataToDoList

		err = rows.Scan(&dataRecordToDoList.ID, &dataRecordToDoList.Name, &dataRecordToDoList.Text, &dataRecordToDoList.DateAdded)
		if err != nil {
			log.Fatal(err)
		}

		// Добавляем объект Data в срез
		dataRecordsToDoList = append(dataRecordsToDoList, dataRecordToDoList)
	}

	// Выполнить запрос на выборку последнего ID
	var lastID int
	err = db.QueryRow("SELECT max(id) FROM todolist").Scan(&lastID)
	if err != nil {
		log.Println(err)
		return
	}

	err = ts.Execute(w, map[string]interface{}{
		"Records": dataRecordsToDoList,
		"ID":      lastID + 1,
	})
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Внутренняя ошибка на сервере2", 500)
		return
	}

}

func formHandlerToDoList(w http.ResponseWriter, r *http.Request) {
	//Проверяем, что метод запроса является POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Используйте r.FormValue для получения значений полей формы
	name := r.FormValue("mainFormsToDoListName")
	text := r.FormValue("mainFormsToDoListText")

	//Вызов функции подключение к БД
	_, err := connectionDb()
	if err != nil {
		log.Println(err)
	}

	//Данные с формы добавляем в БД
	_, err = db.Exec("INSERT INTO todolist (name, text) VALUES (?,?)", name, text)
	if err != nil {
		return
	}

	//Перенаправление
	http.Redirect(w, r, "/todolist", http.StatusSeeOther)

}

// Удалить запись из Списка дел
func deleteEntryToDoList(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		log.Fatal(err)
	}

	//Вызов функции подключение к БД
	_, err = connectionDb()
	if err != nil {
		return
	}

	_, err = db.Exec("DELETE FROM todolist WHERE id = ?", id)
	if err != nil {
		log.Fatal(err)
	}

	http.Redirect(w, r, "/todolist", http.StatusSeeOther)
}

// Редактировать запись из Списка дел
func editPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		log.Fatal(err)
	}
	message := r.URL.Query().Get("message")

	//Вызов функции подключение к БД
	_, err = connectionDb()
	if err != nil {
		log.Println(err)
	}
	_, err = db.Exec("UPDATE todolist SET text = ? WHERE id = ?", message, id)
	if err != nil {
		return
	}

	http.Redirect(w, r, "/todolist", http.StatusSeeOther)
}
