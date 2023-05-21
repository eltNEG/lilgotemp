package schema

type User struct {
	CommonSchema `bson:"inline"`
	Username     string `json:"username,omitempty" bson:"username"`
}

type UserQuery struct {
	Username string `json:"username,omitempty" bson:"username"`
}
