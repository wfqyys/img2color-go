package handler

import (
	"net/http"
)

// Handler Vercel入口函数 - 健康检查
func Handler(w http.ResponseWriter, r *http.Request) {
	healthHandler.ServeHTTP(w, r)
}
