package auth_handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"task-manager/helper"
	auth_dto "task-manager/module/auth/dto"
	auth_service "task-manager/module/auth/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

type authHandler struct {
	service auth_service.AuthService
}

func NewAuthHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	handler := &authHandler{
		service: auth_service.NewAuthService(db),
	}

	routes := group + "/auth"
	{
		router.POST(routes+"/register", handler.Register)
		router.POST(routes+"/login", handler.Login)
		router.GET(routes+"/me", handler.Me)
		router.POST(routes+"/logout", handler.Logout)
	}
}

func (handler *authHandler) Register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req auth_dto.Register
	{
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	req.Name = strings.TrimSpace(req.Name)

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if errs := helper.Validate(req); errs != nil {
		helper.JSON(w, http.StatusBadRequest, map[string]any{"errors": errs})
		return
	}

	if len(req.Password) > 72 {
		helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "password must not exceed 72 bytes"})
		return
	}

	res, err := handler.service.Register(r.Context(), &req)
	{
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	helper.JSON(w, http.StatusCreated, res)
}

func (handler *authHandler) Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req auth_dto.Login
	{
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if errs := helper.Validate(req); errs != nil {
		helper.JSON(w, http.StatusBadRequest, map[string]any{"errors": errs})
		return
	}

	if len(req.Password) > 72 {
		helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "password must not exceed 72 bytes"})
		return
	}

	res, err := handler.service.Login(r.Context(), &req)
	{
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	helper.JSON(w, http.StatusOK, res)
}

func (handler *authHandler) Me(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	token, ok := parseToken(w, r)
	{
		if !ok {
			return
		}
	}

	res, err := handler.service.Authenticate(r.Context(), token)
	{
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	helper.JSON(w, http.StatusOK, res)
}

func (handler *authHandler) Logout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	token, ok := parseToken(w, r)
	{
		if !ok {
			return
		}
	}

	if err := handler.service.Logout(r.Context(), token); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseToken(w http.ResponseWriter, r *http.Request) (string, bool) {
	values := r.Header.Values("Authorization")

	if len(values) != 1 {
		writeServiceError(w, auth_service.ErrInvalidToken)
		return "", false
	}

	parts := strings.Fields(values[0])

	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		writeServiceError(w, auth_service.ErrInvalidToken)
		return "", false
	}

	return parts[1], true
}

func writeServiceError(w http.ResponseWriter, err error) {

	if errors.Is(err, auth_service.ErrEmailExists) {
		helper.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if errors.Is(err, auth_service.ErrInvalidCredentials) || errors.Is(err, auth_service.ErrInvalidToken) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		helper.JSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	if errors.Is(err, auth_service.ErrUserNotActive) {
		helper.JSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}

	log.Printf("Auth service error: %v", err)
	helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}
