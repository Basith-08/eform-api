package database

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

func RunSQLFiles(db *gorm.DB, dirPath string, trackerTable string) error {
	if err := db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (filename TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`, trackerTable)).Error; err != nil {
		return err
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		files = append(files, entry.Name())
	}

	sort.Strings(files)

	for _, fileName := range files {
		var exists int64
		if err := db.Raw(fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE filename = ?", trackerTable), fileName).Scan(&exists).Error; err != nil {
			return err
		}

		if exists > 0 {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dirPath, fileName))
		if err != nil {
			return err
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(string(content)).Error; err != nil {
				return err
			}

			return tx.Exec(fmt.Sprintf("INSERT INTO %s (filename) VALUES (?)", trackerTable), fileName).Error
		}); err != nil {
			return err
		}
	}

	return nil
}
