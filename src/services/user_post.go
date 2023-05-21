package services

import (
	"context"
	"eltneg/goliltemp/src/config"
	"eltneg/goliltemp/src/dependencies"
	"eltneg/goliltemp/src/schema"
)

type CreateUserReq struct {
	Username string `validate:"required" json:"username"`
}

type CreateUserRes struct{}

func (d *CreateUserReq) Controller(ctx context.Context, cfg *config.Config, dpc *dependencies.Dependencies) (status int, msg string, data *CreateUserRes, err error) {
	user := schema.User{Username: d.Username}

	err = dpc.UserCol.Upsert(ctx, &schema.UserQuery{Username: d.Username}, &user)
	if err != nil {
		return 0, "", nil, err
	}

	return 200, "user created", nil, nil
}
