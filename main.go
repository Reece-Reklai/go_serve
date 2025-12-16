package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/Reece-Reklai/go_serve/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	databaseQuery  *database.Queries
	platform       string
	secret         string
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metrics() string {
	return fmt.Sprint(cfg.fileserverHits.Load())
}

func (cfg *apiConfig) resetMetric() {
	cfg.fileserverHits.Swap(int32(0))
}

func main() {
	godotenv.Load(".env")
	var apiCfg apiConfig
	dbURL := os.Getenv("DB_URL")
	apiCfg.platform = os.Getenv("PLATFORM")
	apiCfg.secret = os.Getenv("SECRET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("failed to open database connection")
	}
	dbQueries := database.New(db)
	apiCfg.databaseQuery = dbQueries
	const staticDir = "./public/"
	headerMethod := map[string]string{
		"GET":    "GET",
		"POST":   "POST",
		"PUT":    "PUT",
		"DELETE": "DELETE",
	}
	endPoints := map[string]string{
		"api":   "/api",
		"admin": "/admin",
	}
	router := Router{Mux: http.NewServeMux()}
	const port = "8080"
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router.Mux,
	}
	router.handlers(&apiCfg, staticDir, headerMethod, endPoints)
	log.Fatal(server.ListenAndServe())
}
