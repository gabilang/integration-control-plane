package controllers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/models"
	"github.com/wso2/integration-control-plane/ipaas-service/services"
	"github.com/wso2/integration-control-plane/ipaas-service/utils"
)

// ScheduleController handles HTTP requests for scheduled-task deployment operations.
type ScheduleController interface {
	ListSchedules(w http.ResponseWriter, r *http.Request)
	UpsertSchedule(w http.ResponseWriter, r *http.Request)
	GetSchedule(w http.ResponseWriter, r *http.Request)
	DeleteSchedule(w http.ResponseWriter, r *http.Request)
}

type scheduleController struct {
	service services.ScheduleService
}

func NewScheduleController(service services.ScheduleService) ScheduleController {
	return &scheduleController{service: service}
}

func (c *scheduleController) ListSchedules(w http.ResponseWriter, r *http.Request) {
	org, ok := orgHandle(r)
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	componentName := r.PathValue("componentName")
	projectName := r.URL.Query().Get("projectName")

	list, err := c.service.ListSchedules(r.Context(), org, projectName, componentName)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		slog.ErrorContext(r.Context(), "list schedules failed",
			"error", err, "org", org, "project", projectName, "component", componentName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to list schedules")
		return
	}
	utils.WriteSuccessResponse(w, http.StatusOK, list)
}

func (c *scheduleController) UpsertSchedule(w http.ResponseWriter, r *http.Request) {
	org, ok := orgHandle(r)
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	componentName := r.PathValue("componentName")
	projectName := r.URL.Query().Get("projectName")

	var req models.UpsertScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Environment == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "environment is required")
		return
	}
	if req.CronExpression == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "cronExpression is required")
		return
	}

	schedule, err := c.service.UpsertSchedule(r.Context(), org, projectName, componentName, &req)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		slog.ErrorContext(r.Context(), "upsert schedule failed",
			"error", err, "org", org, "project", projectName, "component", componentName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to upsert schedule")
		return
	}
	utils.WriteSuccessResponse(w, http.StatusOK, schedule)
}

func (c *scheduleController) GetSchedule(w http.ResponseWriter, r *http.Request) {
	org, ok := orgHandle(r)
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	componentName := r.PathValue("componentName")
	environment := r.PathValue("environment")
	projectName := r.URL.Query().Get("projectName")

	schedule, err := c.service.GetSchedule(r.Context(), org, projectName, componentName, environment)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		if errors.Is(err, services.ErrScheduleNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, "schedule not found")
			return
		}
		slog.ErrorContext(r.Context(), "get schedule failed",
			"error", err, "org", org, "project", projectName, "component", componentName, "environment", environment)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get schedule")
		return
	}
	utils.WriteSuccessResponse(w, http.StatusOK, schedule)
}

func (c *scheduleController) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	org, ok := orgHandle(r)
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	componentName := r.PathValue("componentName")
	environment := r.PathValue("environment")
	projectName := r.URL.Query().Get("projectName")

	err := c.service.DeleteSchedule(r.Context(), org, projectName, componentName, environment)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		if errors.Is(err, services.ErrScheduleNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, "schedule not found")
			return
		}
		slog.ErrorContext(r.Context(), "delete schedule failed",
			"error", err, "org", org, "project", projectName, "component", componentName, "environment", environment)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to delete schedule")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
