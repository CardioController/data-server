package main

import (
	"log"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	"CardioController/data-server/helper"
	_ "CardioController/data-server/migrations"
)

func main() {
	app := pocketbase.New()

	// loosely check if it was executed using "go run"
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		// enable auto creation of migration files when making collection changes in the Dashboard
		// (the isGoRun check is to enable it only during development)
		Automigrate: isGoRun,
	})

	app.Cron().MustAdd("Check New Videos", "*/2 * * * *", func() {
		helper.CheckExerciseVideos(app)
		helper.CheckGameplayVideos(app)
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// serves static files from the provided public dir (if exists)
		se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))

		if len(helper.ConfigEnv.ExerciseVideoPath) > 0 {
			se.Router.GET("/exercise_vid/{path...}", apis.Static(os.DirFS(helper.ConfigEnv.ExerciseVideoPath), false))
		}

		if len(helper.ConfigEnv.GamePlayVideoPath) > 0 {
			se.Router.GET("/gameplay_vid/{path...}", apis.Static(os.DirFS(helper.ConfigEnv.GamePlayVideoPath), false))
		}

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
