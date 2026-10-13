package main

import (
	"fmt"
	"net/http"
	"os"
	"task-manager/config"
	"task-manager/helper"
	auth_cmd "task-manager/module/auth"
	category_cmd "task-manager/module/category"
	hudud_cmd "task-manager/module/hudud"

	"github.com/julienschmidt/httprouter"
)

func main() {
	helper.LoadEnv()

	migrate()

	db := config.DBConnect()

	defer db.Close()

	// if len(os.Args) > 1 && os.Args[1] == "db:seed" {

	// 	if len(os.Args) < 3 || os.Args[2] != "categories" {
	// 		fmt.Println("Usage: go run . db:seed categories <count>")
	// 		return
	// 	}

	// 	count := 100

	// 	if len(os.Args) > 3 {
	// 		value, err := strconv.Atoi(os.Args[3])
	// 		{
	// 			if err != nil || value < 1 {
	// 				fmt.Println("Count must be a positive number")
	// 				return
	// 			}
	// 		}

	// 		count = value
	// 	}

	// 	inserted, err := config.SeedCategories(context.Background(), db, count)
	// 	{
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 	}

	// 	fmt.Printf("Categories seeded: %d inserted\n", inserted)

	// 	return
	// }

	router := httprouter.New()
	{
		auth_cmd.Register(router, "/api", db)
		category_cmd.Register(router, "/api", db)
		hudud_cmd.Register(router, "/api", db)
	}

	fmt.Printf("🚀 Server started on http://localhost:%s\n", helper.ENV("APP_PORT"))

	if err := http.ListenAndServe(":"+helper.ENV("APP_PORT"), router); err != nil {
		panic(err)
	}

}

func migrate() {
	if len(os.Args) > 1 {

		switch os.Args[1] {

		case "migrate:create":

			if len(os.Args) < 3 {
				fmt.Println("Usage:   go run . migrate:create <name>")

				fmt.Println("Example: go run . migrate:create add_code_to_regions")

				os.Exit(1)
			}

			config.MigrateCreate(os.Args[2])

			os.Exit(0)

			// case "permissions:push":

			// 	if err := config.PushPermissions(context.Background(), config.AuthURLFromEnv()); err != nil {
			// 		panic(err)
			// 	}

			// 	os.Exit(0)
		}
	}
}
