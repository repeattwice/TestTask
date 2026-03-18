package main

import (
	allnesecary "Test/AllNesecary"
	"context"
	"log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := allnesecary.GetDbConnection(ctx)
	if err != nil {
		log.Fatal("Ошибка подключения:", err)
	}

	if err1 := allnesecary.MigrateTables(db); err1 != nil {
		log.Fatal("Ошибка миграции:", err1)
	}

	allnesecary.CreateServer(*db)
}
