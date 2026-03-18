package allnesecary

import (
	"context"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func GetDbConnection(ctx context.Context) (*gorm.DB, error) {
	connLine := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_port"),
	)
	db, err := gorm.Open(postgres.Open(connLine), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func MigrateTables(db *gorm.DB) error {
	return db.AutoMigrate(&Department{}, &Employees{})

}
