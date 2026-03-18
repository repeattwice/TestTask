package allnesecary

import "time"

type Department struct {
	Id        int       `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:30;not null" json:"name"`
	Parent_id int       `gorm:"index" json:"parent_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type Employees struct {
	Id            int        `gorm:"primaryKey" json:"id"`
	Department_id int        `gorm:"not null" json:"department_id"`
	Full_name     string     `gorm:"size:100;not null" json:"full_name"`
	Position      string     `gorm:"size:50;not null" json:"position"`
	Created_at    time.Time  `gorm:"autoCreateTime" json:"Created_at"`
	Department    Department `gorm:"foreignKey:DepartmentID;constraint:CASCADE,OnDelete:SET NULL;"`
}

type DepartmentResponse struct {
	Department Department           `json:"department"`
	Employees  []Employees          `json:"employees,omitempty"`
	Children   []DepartmentResponse `json:"children,omitempty"`
}

type PatchDto struct {
	Name      *string `json:"name"`
	Parent_id *int    `json:"parent_id"`
}
