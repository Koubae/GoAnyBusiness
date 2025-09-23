package testings

import (
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

func init() {
	currentDir := getCurrentFilePath()
	envPath := filepath.Join(currentDir, "..", "..", "tests", ".env.test")

	err := godotenv.Load(envPath)
	if err != nil {
		panic("Error loading .env.test file: " + err.Error())
	}
}

func getCurrentFilePath() string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Dir(currentFile)
}
