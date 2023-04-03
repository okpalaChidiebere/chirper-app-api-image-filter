package api

import "net/http"


type Interface interface{
	CreateUserHandler() http.HandlerFunc
	SaveTweetHandler() http.HandlerFunc
	ListTweetsHandler() http.HandlerFunc
	ListUsersHandler() http.HandlerFunc
	MigrateTweetsHandler() http.HandlerFunc
	SaveLikeToggleHandler() http.HandlerFunc
}