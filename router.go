package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Reece-Reklai/go_serve/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Router struct {
	Mux *http.ServeMux
}
type SingleChirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
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

func (router *Router) handlers(apiCfg *apiConfig, staticDir string, headerMethod map[string]string, endPoints map[string]string) {
	// App Endpoints ------------------------------------------------------------------------------------------------------

	router.Mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(staticDir)))))

	// Api Endpoints --------------------------------------------------------------------------------------------------------------------------------

	router.Mux.HandleFunc(fmt.Sprintf("%s %s%s", headerMethod["GET"], endPoints["api"], "/healthz"), func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		io.WriteString(w, "OK")
	})

	router.Mux.HandleFunc(fmt.Sprintf("%s %s%s", headerMethod["POST"], endPoints["api"], "/users"), func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		chirpUser := struct {
			Email string `json:"email"`
		}{}
		userJSON := struct {
			ID        uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Email     string    `json:"email"`
		}{}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			respondWithError(w, 400, "failed to read request body")
			return
		}
		err = json.Unmarshal(body, &chirpUser)
		if err != nil {
			respondWithError(w, 400, "failed to unmarshal request")
			return
		}
		user, err := apiCfg.databaseQuery.CreateUser(req.Context(), database.CreateUserParams{ID: uuid.New(), Email: chirpUser.Email})
		if err != nil {
			respondWithError(w, 500, "failed to create user")
			return
		}
		userJSON.ID = user.ID
		userJSON.CreatedAt = user.CreatedAt
		userJSON.UpdatedAt = user.UpdatedAt
		userJSON.Email = user.Email
		response, err := json.Marshal(userJSON)
		if err != nil {
			respondWithError(w, 400, "failed to marshal json into bytes")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		w.Write(response)
	})

	router.Mux.HandleFunc(fmt.Sprintf("%s %s%s", headerMethod["GET"], endPoints["api"], "/chirps"), func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		var responseChirp SingleChirp
		chirps, err := apiCfg.databaseQuery.GetAllChirps(req.Context())
		if err != nil {
			respondWithError(w, 500, "failed to get all chirps")
		}
		allChirps := make([]SingleChirp, 0)
		for _, value := range chirps {
			responseChirp.ID = value.ID
			responseChirp.CreatedAt = value.CreatedAt
			responseChirp.UpdatedAt = value.UpdatedAt
			responseChirp.Body = value.Body
			responseChirp.UserID = value.UserID
			allChirps = append(allChirps, responseChirp)
		}
		respondWithJSON(w, 200, allChirps)
	})
	router.Mux.HandleFunc(fmt.Sprintf("%s %s%s", headerMethod["POST"], endPoints["api"], "/chirps"), func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		chirpProfanity := struct {
			Body   string    `json:"body"`
			UserID uuid.UUID `json:"user_id"`
		}{}
		chirpClean := struct {
			Body   string    `json:"body"`
			UserID uuid.UUID `json:"user_id"`
		}{}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			respondWithError(w, 400, "failed to read request body")
			return
		}

		err = json.Unmarshal(body, &chirpProfanity)
		if err != nil {
			respondWithError(w, 400, "failed to unmarshal request")
			return
		}

		if len(chirpProfanity.Body) > 100 {
			respondWithError(w, 400, "chirp is too long (100 characters)")
			return
		}

		var createWord string
		valid := true
		wordSlice := strings.Split(chirpProfanity.Body, " ")
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
			chirpClean.UserID = chirpProfanity.UserID
			_, err := apiCfg.databaseQuery.CreateChirp(req.Context(), database.CreateChirpParams{Body: chirpClean.Body, UserID: chirpClean.UserID})
			if err != nil {
				respondWithError(w, 500, "failed to create chirp")
			}
			respondWithJSON(w, 201, chirpClean)
			return
		}
		chirpClean.Body = chirpProfanity.Body
		chirpClean.UserID = chirpProfanity.UserID
		_, err = apiCfg.databaseQuery.CreateChirp(req.Context(), database.CreateChirpParams{Body: chirpClean.Body, UserID: chirpClean.UserID})
		if err != nil {
			respondWithError(w, 500, "failed to create chirp")
		}
		respondWithJSON(w, 201, chirpClean)

	})

	// Admin Endpoints --------------------------------------------------------------------------------------------------------------------------------

	router.Mux.HandleFunc(fmt.Sprintf("%s %s%s", headerMethod["GET"], endPoints["admin"], "/metrics"), func(w http.ResponseWriter, req *http.Request) {
		metricHTML := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %v times!</p></body></html>", apiCfg.metrics())
		w.WriteHeader(200)
		io.WriteString(w, metricHTML)
	})

	router.Mux.HandleFunc(fmt.Sprintf("%s %s%s", headerMethod["POST"], endPoints["admin"], "/reset"), func(w http.ResponseWriter, req *http.Request) {
		if apiCfg.platform != "dev" {
			respondWithError(w, 403, "something went wrong")
			return
		}
		apiCfg.resetMetric()
		err := apiCfg.databaseQuery.DeleteAllUsers(req.Context())
		if err != nil {
			respondWithError(w, 500, "something went wrong")
			return
		}
		w.WriteHeader(200)
		io.WriteString(w, "OK")
	})
}
