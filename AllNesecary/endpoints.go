package allnesecary

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// для хранения id департаментов, и валидации сотрудников при создании
var DepIds []int

func CreateDepartment(w http.ResponseWriter, r *http.Request, db *gorm.DB) {

	// получаем json и парсим его в структуру департамента
	var dep Department
	if err := json.NewDecoder(r.Body).Decode(&dep); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// проверяем что департамент, не ссылается сам на себя
	if dep.Id == dep.Parent_id {
		http.Error(w, "Департамент не может являться родителес самого себя", http.StatusNotFound)
		return
	}

	// запоминаем id департамента
	DepIds = append(DepIds, dep.Id)

	//записываем в дб
	result := db.Create(&dep)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	//устанавливем статусы, и выводим созданную структуру
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dep)

}

func CretaEmpl(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	// обрабатываем переменную пути id
	vars := mux.Vars(r)
	idStr := vars["id"]
	DepId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID департамента", http.StatusBadRequest)
		return
	}

	// проверяем, что департамент, в который хотим добавить сотрудника существует
	flag := false
	for i := 0; i < len(DepIds); i++ {
		if DepId == DepIds[i] {
			flag = true
			break
		}
	}
	// обрабатываем случай, когда департамент не существует
	if !flag {
		http.Error(w, "Несуществующий департамент", http.StatusNotFound)
		return
	}

	// получаем json и парсим его в Employees
	var empl Employees
	if err := json.NewDecoder(r.Body).Decode(&empl); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// добавляем id в структуру
	empl.Department_id = DepId

	// записываем структуру в бд
	result := db.Create(&empl)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	//устанавливем статусы, и выводим созданную структуру
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(empl)
}

func GetDepartment(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	// получаем переменную пути и query параметры
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, _ := strconv.Atoi(idStr)
	query := r.URL.Query()
	depth, err := strconv.Atoi(query.Get("depth"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// валидация лишней не будет
	if depth < 1 {
		depth = 1
	} else if depth > 5 {
		depth = 5
	}

	// проверяем надо ли выводить сотрудников
	IncludeEmpl := true
	if IncEmpl := query.Get("include_employees"); IncEmpl == "false" {
		IncludeEmpl = false
	}

	// зарускаем функцию рекусрсивного обращения к бд
	result, err := fetchDepartament(db, id, depth, IncludeEmpl)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Департамерт не найден", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func RotateDepartment(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	// считываем переменные пути
	vars := mux.Vars(r)
	idStr := vars["id"]
	Newid, _ := strconv.Atoi(idStr)

	// испльзую Data transfer object, для удобной проверки, какое значение пришло на вход
	var DTO PatchDto
	if err := json.NewDecoder(r.Body).Decode(&DTO); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	// ищем подразделение, которое надо обновить
	var dept Department
	if err := db.First(&dept, Newid).Error; err != nil {
		http.Error(w, "Подразделение не найдено", http.StatusNotFound)
		return
	}

	// записываем новые значения в мапу
	updates := make(map[string]interface{})
	if DTO.Name != nil {
		updates["name"] = *DTO.Name
	}

	// проверяем, что конкретно пришло в теле запроса, и в заваисимости от запроса, реализуем запрос к бд
	if DTO.Name == nil && DTO.Parent_id != nil {
		if *DTO.Parent_id == Newid {
			http.Error(w, "Подразделение не может быть родителем самого себя", http.StatusBadRequest)
			return
		}
		updates["parent_id"] = *DTO.Parent_id
	}

	if err := db.Model(&dept).Updates(updates).Error; err != nil {
		http.Error(w, "Ошибка при обновлении базы данных", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dept)
}

func DeletDepartment(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	mode := r.URL.Query().Get("mode")
	reassignToIdStr := r.URL.Query().Get("reassign_to_department_id")

	err = db.Transaction(func(tx *gorm.DB) error {
		if mode == "reassign" {
			reassignToId, err := strconv.Atoi(reassignToIdStr)
			if err != nil {
				return fmt.Errorf("reassign_to_department_id is required and must be an integer when mode=reassign")
			}

			var targetDept Department
			if err := tx.First(&targetDept, reassignToId).Error; err != nil {
				return fmt.Errorf("target department not found")
			}

			if err := tx.Model(&Employees{}).Where("department_id = ?", id).Update("department_id", reassignToId).Error; err != nil {
				return err
			}
		} else if mode == "cascade" {
			if err := tx.Where("department_id = ?", id).Delete(&Employees{}).Error; err != nil {
				return err
			}

			if err := tx.Where("parent_id = ?", id).Delete(&Department{}).Error; err != nil {
				return err
			}
		}

		if err := tx.Delete(&Department{}, id).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func fetchDepartament(db *gorm.DB, rootID int, maxDepth int, includeEmployees bool) (DepartmentResponse, error) {
	var allDepts []Department
	if err := db.Find(&allDepts).Error; err != nil {
		return DepartmentResponse{}, err
	}

	var allEmps []Employees
	if includeEmployees {
		db.Find(&allEmps)
	}

	deptMap := make(map[int][]Department)
	for _, d := range allDepts {
		deptMap[d.Parent_id] = append(deptMap[d.Parent_id], d)
	}

	empMap := make(map[int][]Employees)
	for _, e := range allEmps {
		empMap[e.Department_id] = append(empMap[e.Department_id], e)
	}

	return buildTree(rootID, maxDepth, deptMap, empMap, allDepts), nil
}

func buildTree(id int, depth int, deptMap map[int][]Department, empMap map[int][]Employees, allDepts []Department) DepartmentResponse {
	var current Department
	for _, d := range allDepts {
		if d.Id == id {
			current = d
			break
		}
	}

	res := DepartmentResponse{
		Department: current,
		Employees:  empMap[id],
	}

	if depth > 0 {
		for _, childDept := range deptMap[id] {
			res.Children = append(res.Children, buildTree(childDept.Id, depth-1, deptMap, empMap, allDepts))
		}
	}

	return res
}
