package helper

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func readFolder(folderPath string, app *pocketbase.PocketBase, collectionName string) {
	items, _ := os.ReadDir(folderPath)
	for _, item := range items {
		itemAbsolutePath := path.Join(folderPath, item.Name())
		// log.Printf("Processing: %s", itemAbsolutePath)
		if item.IsDir() {
			readFolder(itemAbsolutePath, app, collectionName)
		} else {
			// if item.Name().
			// strings.HasSuffix(item.Name().,"")
			if strings.HasSuffix(strings.ToLower(item.Name()), ".mp4") {
				itemServePath := strings.TrimPrefix(itemAbsolutePath, ConfigEnv.ExerciseVideoPath)
				itemServePath = strings.TrimPrefix(itemServePath, ConfigEnv.GamePlayVideoPath)

				dbItem, _ := app.FindFirstRecordByFilter(collectionName, "file_source_path={:file_source_path}", dbx.Params{
					"file_source_path": itemServePath,
				})
				// if err != nil {
				// 	log.Printf("Error finding existing record in db for [%s]: %v", itemAbsolutePath, err)
				// 	continue
				// }
				if dbItem != nil {
					// log.Printf("DB Record for %s found, skipping this item", itemAbsolutePath)
					continue
				}
				collection, err := app.FindCollectionByNameOrId(collectionName)
				if err != nil {
					log.Printf("Error finding %s collection: %v", collectionName, err)
				}

				record := core.NewRecord(collection)

				record.Set("file_source_path", itemServePath)
				err = app.Save(record)
				if err != nil {
					log.Printf("Error saving new record for item[%s]: %v", itemAbsolutePath, err)
				} else {
					log.Printf("Found new file: [%s]", itemAbsolutePath)
				}
			}
		}
	}
}

func CheckExerciseVideos(app *pocketbase.PocketBase) {
	rawVideoPath := ConfigEnv.ExerciseVideoPath
	if len(rawVideoPath) == 0 {
		log.Print("Raw video path environment variable not defined! Killing job!")
		return
	}
	if exist, err := exists(rawVideoPath); err != nil || !exist {
		log.Print("Raw video path does not exist! Killing job!")
		return
	}

	if stat, err := os.Stat(rawVideoPath); err != nil || !stat.IsDir() {
		log.Print("Raw video path is not a folder! Killing job!")
		return
	}

	readFolder(rawVideoPath, app, "videos_exercise")
}

func CheckGameplayVideos(app *pocketbase.PocketBase) {
	rawVideoPath := ConfigEnv.GamePlayVideoPath
	if len(rawVideoPath) == 0 {
		log.Print("Raw video path environment variable not defined! Killing job!")
		return
	}
	if exist, err := exists(rawVideoPath); err != nil || !exist {
		log.Print("Raw video path does not exist! Killing job!")
		return
	}

	if stat, err := os.Stat(rawVideoPath); err != nil || !stat.IsDir() {
		log.Print("Raw video path is not a folder! Killing job!")
		return
	}

	readFolder(rawVideoPath, app, "videos_gameplay")
}
