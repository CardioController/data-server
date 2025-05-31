package helper

import (
	"log"
	"maps"
	"net/http"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

const EXERCISE_NUM = 3

func GenerateExercise(e *core.RequestEvent, app *pocketbase.PocketBase) error {
	sessionId := e.Request.PathValue("session_id")
	data := struct {
		Category string `json:"category" form:"category"`
	}{}

	if err := e.BindBody(&data); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]bool{"success": false})
	}
	log.Printf("To generate exercise for session [%s], category of exercise: [%s]", sessionId, data.Category)

	session, err := app.FindRecordById("sessions", sessionId)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]bool{"success": false})
	}

	errs := app.ExpandRecord(session, []string{
		"game.metrics_via_game",
		"session_exercises_via_session",
		"videos_gameplay_via_session.gameplay_metric_events_via_video_gameplay.metric",
	}, nil)

	if len(errs) > 0 {
		return e.JSON(http.StatusInternalServerError, map[string]bool{"success": false})
	}
	// log.Println(session.ExpandedAll("game.expand.metrics_via_game"))

	gameRecord := session.ExpandedOne("game")

	metricSummary := map[*core.Record]int{}

	// find all metrics
	for _, metric := range gameRecord.ExpandedAll("metrics_via_game") {
		metricSummary[metric] = 0
	}

	// add all events
	for _, gpVideo := range session.ExpandedAll("videos_gameplay_via_session") {
		for _, gpEvent := range gpVideo.ExpandedAll("gameplay_metric_events_via_video_gameplay") {
			var metric *core.Record
			for k := range maps.Keys(metricSummary) {
				if k.Id == gpEvent.GetString("metric") {
					metric = k
					break
				}
			}
			metricSummary[metric] += 1
		}
	}

	// apply multiplier
	totalSets := 0
	for k := range maps.Keys(metricSummary) {
		metricSummary[k] = metricSummary[k] * k.GetInt("intensity_multiplier")
		totalSets += metricSummary[k] * k.GetInt("intensity_multiplier")
	}

	log.Println(metricSummary)

	// remove old session_exercises records for this session if any
	for _, oldExercise := range session.ExpandedAll("session_exercises_via_session") {
		app.Delete(oldExercise)
	}

	// randomly choose exercise
	exercises, err := app.FindRecordsByFilter("exercises", "categories~{:category}", "@random", EXERCISE_NUM, 0, dbx.Params{
		"category": data.Category,
	})

	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]bool{"success": false})
	}

	// create exercise records
	sessionExerciseCollection, err := app.FindCollectionByNameOrId("session_exercises")
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]bool{"success": false})
	}
	for idx, ex := range exercises {
		record := core.NewRecord(sessionExerciseCollection)
		record.Set("exercise_order", idx)
		record.Set("session", sessionId)
		record.Set("exercise", ex.Id)
		// record.Set("sets", totalSets)
		err = app.Save(record)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]bool{"success": false})
		}
	}

	session.Set("exercise_sets", totalSets)
	app.Save(session)

	log.Printf("Found exercises: %v", exercises)

	return e.JSON(http.StatusOK, map[string]bool{"success": true})
}
