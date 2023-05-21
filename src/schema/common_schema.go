package schema

import "time"

type CommonSchema struct {
	CreatedAt int64 `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt int64 `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}

func (c *CommonSchema) SetCreatedAt() {
	if c.CreatedAt == 0 {
		c.CreatedAt = time.Now().UnixMilli()
	}

}
func (c *CommonSchema) SetUpdatedAt() {
	c.UpdatedAt = time.Now().UnixMilli()
}
