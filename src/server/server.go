package server

import (
	"eltneg/goliltemp/src/config"
	"eltneg/goliltemp/src/db"
	"eltneg/goliltemp/src/dbsetup"
	"eltneg/goliltemp/src/dependencies"
	"eltneg/goliltemp/src/models"
	"eltneg/goliltemp/src/schema"
	"eltneg/goliltemp/src/services"
	"eltneg/goliltemp/src/utils"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	log "github.com/sirupsen/logrus"
)

func Run() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	cfg, err := config.Init()
	if err != nil {
		log.Println("error loading config", err)
		return
	}

	_, Db, err := db.Init(cfg, dbsetup.DBindexes, nil)
	if err != nil {
		log.Fatal("Failed to init db: ", err)
	}

	d := &dependencies.Dependencies{
		UserCol: models.DBModel[*schema.User, *schema.UserQuery]{Store: &Db, Name: "address"},
	}

	ping := (&utils.H[*services.PingReq, *services.PingRes]{}).Handler(cfg, nil, &services.PingReq{})
	createUser := (&utils.H[*services.CreateUserReq, *services.CreateUserRes]{}).Handler(cfg, d, &services.CreateUserReq{})

	r.Get("/ping", ping)
	r.Post("/user", createUser)

	log.Info("server started")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), r))
}
