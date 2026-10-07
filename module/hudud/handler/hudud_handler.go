package hudud_handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"task-manager/helper"
	hudud_dto "task-manager/module/hudud/dto"
	hudud_service "task-manager/module/hudud/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

type hududHandler struct {
	service hudud_service.HududService
}

func NewHududHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	handler := &hududHandler{
		service: hudud_service.NewHududService(db),
	}

	routes := group + "/hududs"
	{
		router.POST(routes, handler.Create)
		router.PUT(routes+"/:id", handler.Update)
		router.DELETE(routes+"/:id", handler.Delete)
		router.GET(routes+"/:id", handler.Show)
		router.GET(routes, handler.List)
	}
}

func (handler *hududHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req hudud_dto.Create
	{
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	if errs := helper.Validate(req); errs != nil {
		helper.JSON(w, http.StatusBadRequest, map[string]any{"errors": errs})
		return
	}

	res, err := handler.service.Create(r.Context(), &req)
	{
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	helper.JSON(w, http.StatusCreated, res)
}

func (handler *hududHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, ok := parseID(w, ps)
	{
		if !ok {
			return
		}
	}

	var req hudud_dto.Update
	{
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	if errs := helper.Validate(req); errs != nil {
		helper.JSON(w, http.StatusBadRequest, map[string]any{"errors": errs})
		return
	}

	res, err := handler.service.Update(r.Context(), id, &req)
	{
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	helper.JSON(w, http.StatusOK, res)
}

func (handler *hududHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, ok := parseID(w, ps)
	{
		if !ok {
			return
		}
	}

	if err := handler.service.Delete(r.Context(), id); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (handler *hududHandler) Show(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, ok := parseID(w, ps)
	{
		if !ok {
			return
		}
	}

	res, err := handler.service.Show(r.Context(), id)
	{
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	helper.JSON(w, http.StatusOK, res)
}

func (handler *hududHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	q := r.URL.Query()

	filter := &hudud_dto.Filter{
		Name:      q.Get("name"),
		Direction: "next",
		Limit:     20,
	}

	if q.Get("direction") != "" {

		if q.Get("direction") != "next" && q.Get("direction") != "prev" {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "direction must be next or prev"})
			return
		}

		filter.Direction = q.Get("direction")
	}

	if q.Get("id") != "" {

		value, err := strconv.ParseInt(q.Get("id"), 10, 64)
		{
			if err != nil || value < 1 {
				helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
				return
			}
		}

		filter.ID = value
	}

	if q.Get("is_active") != "" {

		value, err := strconv.ParseBool(q.Get("is_active"))
		{
			if err != nil {
				helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid is_active"})
				return
			}
		}

		filter.IsActive = &value
	}

	if q.Get("cursor") != "" {

		value, err := strconv.ParseInt(q.Get("cursor"), 10, 64)
		{
			if err != nil || value < 1 {
				helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid cursor"})
				return
			}
		}

		filter.Cursor = &value
	}

	if filter.Direction == "prev" && filter.Cursor == nil {
		helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "cursor is required for prev direction"})
		return
	}

	if q.Get("limit") != "" {

		value, err := strconv.Atoi(q.Get("limit"))
		{
			if err != nil || value < 1 || value > 100 {
				helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid limit"})
				return
			}
		}

		filter.Limit = value
	}

	res, err := handler.service.List(r.Context(), filter)
	{
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	helper.JSON(w, http.StatusOK, res)
}

func parseID(w http.ResponseWriter, ps httprouter.Params) (int64, bool) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	{
		if err != nil || id < 1 {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
			return 0, false
		}
	}

	return id, true
}

func writeServiceError(w http.ResponseWriter, err error) {

	if errors.Is(err, pgx.ErrNoRows) {
		helper.JSON(w, http.StatusNotFound, map[string]string{"error": "hudud not found"})
		return
	}

	helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
}
