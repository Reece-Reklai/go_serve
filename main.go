package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/Reece-Reklai/go_serve/internal"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
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

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
	return nil
}

func respondWithError(w http.ResponseWriter, code int, msg string) error {
	return respondWithJSON(w, code, map[string]string{"error": msg})
}

func main() {
	// dbURL := os.Getenv("DB_URL")
	// db, err := sql.Open("postgres", dbURL)
	// if err != nil {
	// 	fmt.Println("failed to open database connection")
	// }
	// dbQueries := database.New(db)
	var apiCfg apiConfig
	staticDir := "./public/"
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
	router := internal.Router{Mux: http.NewServeMux()}
	port := "8080"
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router.Mux,
	}
	router.Mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(staticDir)))))
	router.Mux.HandleFunc(fmt.Sprintf("%s %s%s", headerMethod["GET"], endPoints["api"], "/healthz"), func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		io.WriteString(w, "OK")
	})
	router.Mux.HandleFunc(fmt.Sprintf("%s %s%s", headerMethod["POST"], endPoints["api"], "/validate_chirp"), func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		chirpBody := struct {
			Body string `json:"body"`
		}{}
		chirpClean := struct {
			Body string `json:"cleaned_body"`
		}{}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			respondWithError(w, 400, "could not read request")
			return
		}

		err = json.Unmarshal(body, &chirpBody)
		if err != nil {
			respondWithError(w, 400, "could not unmarshal request")
			return
		}

		if len(chirpBody.Body) > 140 {
			respondWithError(w, 400, "chirp is too long")
			return
		}

		var createWord string
		valid := true
		wordSlice := strings.Split(chirpBody.Body, " ")
		for index := range wordSlice {
			switch strings.ToLower(wordSlice[index]) {
			case "kerfuffle":
				valid = false
				wordSlice[index] = "****"
			case "sharbert":
				valid = false
				wordSlice[index] = "****"
			case "fornax":
				valid = false
				wordSlice[index] = "****"
			}
		}
		for index, val := range wordSlice {
			if index == 0 {
				createWord = val
				continue
			}
			createWord = createWord + " " + val
		}
		if valid == false {
			chirpClean.Body = createWord
			respondWithJSON(w, 200, chirpClean)
			return
		}
		chirpClean.Body = chirpBody.Body
		respondWithJSON(w, 200, chirpClean)

	})
	router.Mux.HandleFunc(fmt.Sprintf("%s %s%s", headerMethod["GET"], endPoints["admin"], "/metrics"), func(w http.ResponseWriter, req *http.Request) {
		metricHTML := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %v times!</p></body></html>", apiCfg.metrics())
		w.WriteHeader(200)
		io.WriteString(w, metricHTML)
	})
	router.Mux.HandleFunc(fmt.Sprintf("%s %s%s", headerMethod["POST"], endPoints["admin"], "/reset"), func(w http.ResponseWriter, req *http.Request) {
		apiCfg.resetMetric()
		w.WriteHeader(200)
	})
	log.Fatal(server.ListenAndServe())
}
