//Filename: cmd/api/healthcheck.go

package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	//js := `{"status": "available", "environment": %q, "version": %q }`
	//js = fmt.Sprintf(js, app.config.env, version)

	//Create a map to hold our healthcheck data
	data := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}
	// //simulate a delay
	// time.Sleep(4 * time.Second)

	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
