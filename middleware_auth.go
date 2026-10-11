package main

import (
	"fmt"
	"net/http"

	"github.com/anuj-jha-work/rss_Agregator/internal/auth"
	database "github.com/anuj-jha-work/rss_Agregator/internal/database"
)

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func (cfg *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apikey, err := auth.GetApikey(r.Header)
		if err != nil {
			respondWithError(w, 403, fmt.Sprintf("Error getting API key: %v", err))
			return
		}
		user, err := cfg.DB.GetUserByApiKey(r.Context(), apikey)
		if err != nil {
			respondWithError(w, 400, fmt.Sprintf("Couldn't get user by API key: %v", err))
			return
		}
		handler(w, r, user)
	}
}
