package master

import (
	"net/http"
	"time"
)

type Handler struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	StartTime      time.Time
}

func NewHandler(w http.ResponseWriter, r *http.Request) Handler {
	return Handler{w, r, time.Now()}
}

type HandlerFunc func(Handler)
