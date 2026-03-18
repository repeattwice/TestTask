package allnesecary

import (
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func CreateServer(db gorm.DB) {
	router := mux.NewRouter()
	port := "5050"
	router.HandleFunc("/departments", func(w http.ResponseWriter, r *http.Request) {
		CreateDepartment(w, r, &db)
	}).Methods("POST") //Создать подразделение

	router.HandleFunc("/departments/{id}/employees", func(w http.ResponseWriter, r *http.Request) {
		CretaEmpl(w, r, &db)
	}).Methods("POST") //Создать сотрудника в подразделении

	router.HandleFunc("/departments/{id}", func(w http.ResponseWriter, r *http.Request) {
		GetDepartment(w, r, &db)
	}).Methods("GET") // Получить подразделение

	router.HandleFunc("/departments/{id}", func(w http.ResponseWriter, r *http.Request) {
		RotateDepartment(w, r, &db)
	}).Methods("PATCH") //Переместить подразделение в другое

	router.HandleFunc("/departments", func(w http.ResponseWriter, r *http.Request) {
		DeletDepartment(w, r, &db)
	}).Methods("DELETE") //Удалить подразделение

	http.ListenAndServe(":"+port, router)
}
