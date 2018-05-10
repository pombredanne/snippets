package util

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// RegisterHealthPage adds a "/health" page to a router that only
// returns success when database connectivity works properly.
func RegisterHealthPage(database *gorm.DB, router *mux.Router) {
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if rows, err := database.DB().Query("SELECT 1"); err == nil {
			rows.Close()
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}
