package helper

import (
	"os"
)

type AppConfigEnv struct {
	Debug             bool
	ExerciseVideoPath string
	GamePlayVideoPath string
}

var ConfigEnv *AppConfigEnv

func init() {
	ConfigEnv = &AppConfigEnv{
		ExerciseVideoPath: os.Getenv("EXERCISE_VID_PATH"),
		GamePlayVideoPath: os.Getenv("GAMEPLAY_VID_PATH"),
		Debug:             os.Getenv("DEBUG") != "1",
	}
}
