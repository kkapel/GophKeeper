package server

import "github.com/kkapel/gophkeeper/internal/db/connections"

func Run() error {
	var dbstr string = ""

	// Инициализация БД
	db, err := connections.InitDB(dbstr, "./migrations")
	if err != nil {
		return err
	}

	// Инициализация сервиса
	//gophKeeper := &GophKeeper{
	//	db: db,
	//	}
	// Закрываем БД-соединение
	if db != nil {
		defer func() { _ = db.Close() }()
	}

	return nil
}
