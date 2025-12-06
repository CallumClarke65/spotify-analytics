package services

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
)

func WriteJsonObjectToFile(obj interface{}, filename string) error {

	// Ensure the "docs" directory exists
	err := os.MkdirAll("./docs", os.ModePerm)

	if err != nil {
		zap.L().Error("failed to create docs directory", zap.Error(err))
		return err
	}

	// Create or overwrite the file
	if !strings.HasSuffix(filename, ".json") {
		filename = fmt.Sprintf("%s.json", filename)
	}
	path := fmt.Sprintf("./docs/%s", filename)
	file, err := os.Create(path)
	if err != nil {
		zap.L().Error("failed to save object", zap.Error(err))
		return err
	}
	defer file.Close()

	// Configure JSON encoder with indentation
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	enc.Encode(obj)

	zap.L().Info("saved object JSON to file", zap.String("path", path))

	return nil
}
