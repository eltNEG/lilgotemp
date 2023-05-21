package dependencies

import (
	"eltneg/goliltemp/src/models"
	"eltneg/goliltemp/src/schema"
)

type Dependencies struct {
	UserCol models.DBModel[*schema.User, *schema.UserQuery]
}
