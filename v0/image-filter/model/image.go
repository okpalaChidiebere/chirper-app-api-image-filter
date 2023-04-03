package imagefiltermodel

import (
	"time"
)

type Image struct {
	Url string  `json:"url" dynamodbav:"author"` //the url must be unique
	CreatedAt time.Time `json:"created_at" dynamodbav:"created_at,unixtime"`
  	UpdatedAt time.Time  `json:"updated_at" dynamodbav:"updated_at,unixtime"` //if empty then we know its a new tweet
}
