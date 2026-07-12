package client

import (
	"os"
	"path/filepath"
)

// tokenFilePath возвращает путь к файлу с токеном: ~/.gophkeeper/token
func tokenFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gophkeeper", "token"), nil
}

// saveToken сохраняет токен в файл в домашней директории пользователя.
func saveToken(token string) error {
	path, err := tokenFilePath()
	if err != nil {
		return err
	}

	// Создаём директорию, если её нет
	// 0700 — только владелец может читать, писать и выполнять
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(token), 0600) // 0600 — только владелец может читать и писать
}

// loadToken читает сохранённый токен из файла.
func loadToken() (string, error) {
	path, err := tokenFilePath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
