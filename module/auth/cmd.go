package auth_cmd

import (
	auth_handler "task-manager/module/auth/handler"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

func Register(router *httprouter.Router, group string, db *pgxpool.Pool) {
	auth_handler.NewAuthHandler(router, group, db)
}
