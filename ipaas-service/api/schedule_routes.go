package api

import (
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/controllers"
)

func registerScheduleRoutes(mux *http.ServeMux, c controllers.ScheduleController) {
	mux.HandleFunc("GET /components/{componentName}/schedules", c.ListSchedules)
	mux.HandleFunc("POST /components/{componentName}/schedules", c.UpsertSchedule)
	mux.HandleFunc("GET /components/{componentName}/schedules/{environment}", c.GetSchedule)
	mux.HandleFunc("DELETE /components/{componentName}/schedules/{environment}", c.DeleteSchedule)
}
