package hudud_cmd

import (
	hudud_handler "task-manager/module/hudud/handler"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

func Register(router *httprouter.Router, group string, db *pgxpool.Pool) {
	hudud_handler.NewHududHandler(router, group, db)
}
