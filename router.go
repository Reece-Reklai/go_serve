package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Reece-Reklai/go_serve/internal/auth"
	"github.com/Reece-Reklai/go_serve/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Router struct {
	Mux *http.ServeMux
}

type userHandler struct {
	post          string
	put           string
	secret        string
	databaseQuery *database.Queries
}

func (userHandler *userHandler) create(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	getUSER := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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
	err = json.Unmarshal(body, &getUSER)
	if err != nil {
		respondWithError(w, 400, "failed to unmarshal request")
		return
	}
	hashPassword, err := auth.HashPassword(getUSER.Password)
	if err != nil {
		respondWithError(w, 400, "failed to hash")
		return
	}
	user, err := userHandler.databaseQuery.CreateUser(req.Context(), database.CreateUserParams{ID: uuid.New(), Email: getUSER.Email, Password: hashPassword})
	if err != nil {
		error := fmt.Sprintf("failed: %v", err)
		respondWithError(w, 500, error)
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
}

func (userHandler *userHandler) update(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	userJSON := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	bearerToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		error := fmt.Sprintf("failed: %v", err)
		respondWithError(w, 401, error)
		return
	}
	userID, err := auth.ValidateJWT(bearerToken, userHandler.secret)
	if err != nil {
		error := fmt.Sprintf("failed: %v", err)
		respondWithError(w, 401, error)
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		respondWithError(w, 400, "failed to read request body")
		return
	}
	err = json.Unmarshal(body, &userJSON)
	if err != nil {
		respondWithError(w, 400, "failed to unmarshal request")
		return
	}
	hashPassword, err := auth.HashPassword(userJSON.Password)
	if err != nil {
		error := fmt.Sprintf("failed: %v", err)
		respondWithError(w, 500, error)
		return
	}
	err = userHandler.databaseQuery.UpdateUserByEmailAndPassword(req.Context(), database.UpdateUserByEmailAndPasswordParams{Email: userJSON.Email, Password: hashPassword, ID: userID})
	if err != nil {
		error := fmt.Sprintf("failed: %v", err)
		respondWithError(w, 500, error)
		return
	}
	respondWithJSON(w, 200, userJSON)
}

type loginHandler struct {
	post          string
	secret        string
	databaseQuery *database.Queries
}

func (loginHandler *loginHandler) create(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	authenticateUser := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	userJSON := struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}{}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		respondWithError(w, 400, "failed to read request body")
		return
	}
	err = json.Unmarshal(body, &authenticateUser)
	if err != nil {
		respondWithError(w, 400, "failed to unmarshal request")
		return
	}
	user, err := loginHandler.databaseQuery.GetUserByEmail(req.Context(), authenticateUser.Email)
	if err != nil {
		error := fmt.Sprintf("failed: %v", err)
		respondWithError(w, 500, error)
		return
	}
	match, err := auth.CheckPasswordHash(authenticateUser.Password, user.Password)
	if err != nil {
		respondWithError(w, 500, "failed to check password match")
		return
	}
	if match == true {
		expire := time.Duration(1) * time.Hour
		jwt, err := auth.MakeJWT(user.ID, loginHandler.secret, expire)
		if err != nil {
			respondWithError(w, 500, "invalid jwt creation")
			return
		}
		refreshToken, err := auth.MakeRefreshToken()
		if err != nil {
			respondWithError(w, 500, "invalid refresh token creation")
			return
		}
		expireAccessToken := time.Duration(1440) * time.Hour
		_, err = loginHandler.databaseQuery.CreateToken(req.Context(), database.CreateTokenParams{ID: refreshToken, ExpiresAt: time.Now().Add(expireAccessToken), UserID: user.ID})
		if err != nil {
			error := fmt.Sprintf("failed: %v", err)
			respondWithError(w, 500, error)
			return
		}
		userJSON.Token = jwt
		userJSON.RefreshToken = refreshToken
		userJSON.ID = user.ID
		userJSON.CreatedAt = user.CreatedAt
		userJSON.UpdatedAt = user.UpdatedAt
		userJSON.Email = user.Email
		respondWithJSON(w, 200, userJSON)
		return
	}
}

type chirpHandler struct {
	post          string
	deleteChirp   string
	getAllChirps  string
	getChirpID    string
	secret        string
	databaseQuery *database.Queries
}

func (chirpHandler *chirpHandler) create(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	chirpProfanity := struct {
		ID     uuid.UUID `json:"id"`
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}{}
	chirpClean := struct {
		ID     uuid.UUID `json:"id"`
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}{}
	bearerToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		error := fmt.Sprintf("failed: %v", err)
		respondWithError(w, 401, error)
		return
	}
	validToken, err := auth.ValidateJWT(bearerToken, chirpHandler.secret)
	if err != nil {
		error := fmt.Sprintf("failed: %v", err)
		respondWithError(w, 401, error)
		return
	}
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
		chirpClean.UserID = validToken
		chirp, err := chirpHandler.databaseQuery.CreateChirp(req.Context(), database.CreateChirpParams{Body: chirpClean.Body, UserID: validToken})
		chirpClean.ID = chirp.ID
		if err != nil {
			error := fmt.Sprintf("failed: %v", err)
			respondWithError(w, 500, error)
			return
		}
		respondWithJSON(w, 201, chirpClean)
		return
	} else {
		chirpClean.Body = chirpProfanity.Body
		chirpClean.UserID = validToken
		chirp, err := chirpHandler.databaseQuery.CreateChirp(req.Context(), database.CreateChirpParams{Body: chirpClean.Body, UserID: validToken})
		chirpClean.ID = chirp.ID
		if err != nil {
			error := fmt.Sprintf("failed: %v", err)
			respondWithError(w, 500, error)
			return
		}
		respondWithJSON(w, 201, chirpClean)
		return
	}
}

type ChirpRow struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (chirpHandler *chirpHandler) getAllChirp(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	var responseChirp ChirpRow
	chirps, err := chirpHandler.databaseQuery.GetAllChirps(req.Context())
	if err != nil {
		error := fmt.Sprintf("failed: %v", err)
		respondWithError(w, 500, error)
		return
	}
	allChirps := make([]ChirpRow, 0)
	for _, value := range chirps {
		responseChirp.ID = value.ID
		responseChirp.CreatedAt = value.CreatedAt
		responseChirp.UpdatedAt = value.UpdatedAt
		responseChirp.Body = value.Body
		responseChirp.UserID = value.UserID
		allChirps = append(allChirps, responseChirp)
	}
	respondWithJSON(w, 200, allChirps)
}

func (chirpHandler *chirpHandler) getChirpByID(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	var responseChirp ChirpRow
	chirpID := req.PathValue("chirpID")
	if chirpID == "" {
		respondWithError(w, 404, "no match found in request")
		return
	}
	stringToUUID, err := uuid.ParseBytes([]byte(chirpID))
	if err != nil {
		respondWithError(w, 400, "failed to parse request")
		return
	}
	chirp, err := chirpHandler.databaseQuery.GetChirpById(req.Context(), stringToUUID)
	if err != nil {
		error := fmt.Sprintf("failed: %v", err)
		respondWithError(w, 404, error)
		return
	}
	responseChirp.ID = chirp.ID
	responseChirp.CreatedAt = chirp.CreatedAt
	responseChirp.UpdatedAt = chirp.UpdatedAt
	responseChirp.Body = chirp.Body
	responseChirp.UserID = chirp.UserID
	respondWithJSON(w, 200, responseChirp)
}

func (chirpHandler *chirpHandler) delete(w http.ResponseWriter, req *http.Request) {
	req.Body.Close()
	bearerToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		error := fmt.Sprintf("failed: %v", err)
		respondWithError(w, 401, error)
		return
	}
	tokenSubjectUserID, err := auth.ValidateJWT(bearerToken, chirpHandler.secret)
	if err != nil {
		error := fmt.Sprintf("failed: %v", err)
		respondWithError(w, 401, error)
		return
	}
	chirpID := req.PathValue("chirpID")
	if chirpID == "" {
		respondWithError(w, 404, "no match found in request")
		return
	}
	stringToUUID, err := uuid.ParseBytes([]byte(chirpID))
	if err != nil {
		respondWithError(w, 500, "failed to parse request")
		return
	}
	chirp, err := chirpHandler.databaseQuery.GetChirpById(req.Context(), stringToUUID)
	if err != nil {
		error := fmt.Sprintf("failed: %v", err)
		respondWithError(w, 404, error)
		return
	}
	if chirp.UserID != tokenSubjectUserID {
		error := fmt.Sprintf("failed: %v", err)
		respondWithError(w, 403, error)
		return
	}
	err = chirpHandler.databaseQuery.DeleteChirpByUserID(req.Context(), database.DeleteChirpByUserIDParams{ID: stringToUUID, UserID: tokenSubjectUserID})
	if err != nil {
		error := fmt.Sprintf("failed: %v", err)
		respondWithError(w, 404, error)
		return
	}
	w.WriteHeader(204)
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
	postUser := fmt.Sprintf("%s %s%s", headerMethod["POST"], endPoints["api"], "/users")
	postLogin := fmt.Sprintf("%s %s%s", headerMethod["POST"], endPoints["api"], "/login")
	postChirp := fmt.Sprintf("%s %s%s", headerMethod["POST"], endPoints["api"], "/chirps")
	postRefresh := fmt.Sprintf("%s %s%s", headerMethod["POST"], endPoints["api"], "/refresh")
	postRevoke := fmt.Sprintf("%s %s%s", headerMethod["POST"], endPoints["api"], "/revoke")
	getAllChirps := fmt.Sprintf("%s %s%s", headerMethod["GET"], endPoints["api"], "/chirps")
	getChirpID := fmt.Sprintf("%s %s%s", headerMethod["GET"], endPoints["api"], "/chirps/{chirpID}")
	putUser := fmt.Sprintf("%s %s%s", headerMethod["PUT"], endPoints["api"], "/users")
	deleteChirp := fmt.Sprintf("%s %s%s", headerMethod["DELETE"], endPoints["api"], "/chirps/{chirpID}")
	user := userHandler{post: postUser, put: putUser, secret: apiCfg.secret, databaseQuery: apiCfg.databaseQuery}
	login := loginHandler{secret: apiCfg.secret, post: postLogin, databaseQuery: apiCfg.databaseQuery}
	chirp := chirpHandler{secret: apiCfg.secret, post: postChirp, getAllChirps: getAllChirps, deleteChirp: deleteChirp, getChirpID: getChirpID, databaseQuery: apiCfg.databaseQuery}

	// App Endpoints ------------------------------------------------------------------------------------------------------

	router.Mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(staticDir)))))

	// Api Endpoints --------------------------------------------------------------------------------------------------------------------------------

	// POST methods
	router.Mux.HandleFunc(user.post, user.create)
	router.Mux.HandleFunc(login.post, login.create)
	router.Mux.HandleFunc(chirp.post, chirp.create)
	router.Mux.HandleFunc(postRefresh, func(w http.ResponseWriter, req *http.Request) {
		token := struct {
			Token string `json:"token"`
		}{}
		BearerToken, err := auth.GetBearerToken(req.Header)
		if err != nil {
			respondWithError(w, 400, "invalid request header token")
			return
		}
		refreshToken, err := apiCfg.databaseQuery.GetTokenByID(req.Context(), BearerToken)
		if err != nil {
			error := fmt.Sprintf("failed: %v", err)
			respondWithError(w, 401, error)
			return
		}
		if time.Now().After(refreshToken.ExpiresAt) {
			respondWithError(w, 401, "refresh token expired")
			return
		}
		if refreshToken.RevokedAt.Valid {
			respondWithError(w, 401, "refresh token revoked")
			return
		}
		expire := time.Duration(1) * time.Hour
		newAccessToken, err := auth.MakeJWT(refreshToken.UserID, apiCfg.secret, expire)
		if err != nil {
			respondWithError(w, 500, "invalid access token creation")
			return
		}
		token.Token = newAccessToken
		respondWithJSON(w, 200, token)
	})
	router.Mux.HandleFunc(postRevoke, func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		BearerToken, err := auth.GetBearerToken(req.Header)
		if err != nil {
			respondWithError(w, 400, "invalid request header token")
			return
		}
		refreshToken, err := apiCfg.databaseQuery.GetTokenByID(req.Context(), BearerToken)
		if err != nil {
			error := fmt.Sprintf("failed: %v", err)
			respondWithError(w, 401, error)
			return
		}
		err = apiCfg.databaseQuery.RevokeToken(req.Context(), database.RevokeTokenParams{ExpiresAt: time.Now(), UpdatedAt: time.Now(), ID: refreshToken.ID})
		if err != nil {
			error := fmt.Sprintf("failed: %v", err)
			respondWithError(w, 401, error)
			return
		}
		w.WriteHeader(204)
	})

	// PUT methods
	router.Mux.HandleFunc(user.put, user.update)

	// DELETE methods
	router.Mux.HandleFunc(chirp.deleteChirp, chirp.delete)

	// GET methods
	router.Mux.HandleFunc(chirp.getChirpID, chirp.getChirpByID)
	router.Mux.HandleFunc(chirp.getAllChirps, chirp.getAllChirp)
	router.Mux.HandleFunc(fmt.Sprintf("%s %s%s", headerMethod["GET"], endPoints["api"], "/healthz"), func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		io.WriteString(w, "OK")
	})

	// Admin Endpoints --------------------------------------------------------------------------------------------------------------------------------

	// POST methods
	router.Mux.HandleFunc(fmt.Sprintf("%s %s%s", headerMethod["POST"], endPoints["admin"], "/reset"), func(w http.ResponseWriter, req *http.Request) {
		if apiCfg.platform != "dev" {
			respondWithError(w, 403, "something went wrong permission rights")
			return
		}
		apiCfg.resetMetric()
		err := apiCfg.databaseQuery.DeleteAllUsers(req.Context())
		if err != nil {
			respondWithError(w, 500, "something went wrong with database")
			return
		}
		w.WriteHeader(200)
		io.WriteString(w, "OK")
	})

	// GET methods
	router.Mux.HandleFunc(fmt.Sprintf("%s %s%s", headerMethod["GET"], endPoints["admin"], "/metrics"), func(w http.ResponseWriter, req *http.Request) {
		metricHTML := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %v times!</p></body></html>", apiCfg.metrics())
		w.WriteHeader(200)
		io.WriteString(w, metricHTML)
	})
}
