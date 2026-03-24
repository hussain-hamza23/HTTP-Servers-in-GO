package main

import (
	"database/sql"
	"encoding/json"
	"httpserver/internal/database"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type filePaths struct {
	app string
	assets string
}


type TokenSecret string
func (ts TokenSecret) String() string{
	return ""
}

type PolkaKey string
func (pk PolkaKey) String() string{
	return ""
}

type apiConfig struct{
	fileserverHits atomic.Int64
	dbQueries *database.Queries
	tokenSecret TokenSecret
	polkaKey PolkaKey
	role string
}

func getHandlers(mux *http.ServeMux, fileDirs *filePaths, cfg *apiConfig){
	mux.Handle("/app/", cfg.middlewareMetricsIncrement(http.StripPrefix("/app", http.FileServer(http.Dir(fileDirs.app)))))
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(fileDirs.assets))))
	//Metrics
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", cfg.numberofHits)
	mux.HandleFunc("POST /admin/reset", cfg.resetUsersHandler)

	//Create requests
	mux.HandleFunc("POST /api/chirps", cfg.createChirpHandler)
	mux.HandleFunc("POST /api/users", cfg.createUserHandler)
	//Retreive requests
	mux.HandleFunc("GET /api/chirps", cfg.getChirpsHandler)
	//mux.HandleFunc("GET /api/chirps/{id}", cfg.getSingleChirpHandler)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.deleteChirpHandler)
	mux.HandleFunc("POST /api/login", cfg.loginHandler)
	mux.HandleFunc("POST /api/refresh", cfg.refreshTokensHandler)
	mux.HandleFunc("POST /api/revoke", cfg.revokeTokenHandler)
	mux.HandleFunc("PUT /api/users",cfg.updateUserPasswordHandler)

	mux.HandleFunc("POST /api/polka/webhooks", cfg.upgradeToChirpyRedHandler)
}

func handlerReadiness(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		http.Error(w, "Status is not OK", http.StatusInternalServerError)
		return
	}

}

func decodeRequest[T any](w http.ResponseWriter, req *http.Request) (T, *HTTPError){
	var payload T
	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()
	if err := decoder.Decode(&payload); err != nil{
		httpError := &HTTPError{
			StatusCode: http.StatusBadRequest,
			Message: ErrDecodingRequest,
			Err: err,
		}
		httpError.errorResponse(w)
		return payload, httpError
	}
	return payload, nil
}


func main(){
	godotenv.Load()
	var dbURL string = os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is not set")
	}

	var platform string = os.Getenv("PLATFORM")
	var tokenSecret TokenSecret = TokenSecret(os.Getenv("SECRET"))
	if tokenSecret == "" {
		log.Fatal("SECRET environment variable is not set")
	}
	var polkaKey PolkaKey = PolkaKey(os.Getenv("POLKA_KEY"))
	if polkaKey == "" {
		log.Fatal("POLKA_KEY environment variable is not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %s", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping the database: %s", err)
	}
	//defer db.Close()
	
	const addr string = ":8080"
	var cfg apiConfig = apiConfig{
		fileserverHits: atomic.Int64{},
		dbQueries: database.New(db),
		tokenSecret: tokenSecret,
		polkaKey: polkaKey,
		role: platform,
	}
	var fileDirs filePaths = filePaths{
		app: "./app/",
		assets: "./assets/",
	}

	var mux *http.ServeMux = http.NewServeMux()

	getHandlers(mux, &fileDirs, &cfg)

	var srv *http.Server = &http.Server{
		Handler: mux,
		Addr: addr,
	}
	
	log.Printf("Serving files from %s on port %s\n", fileDirs.app, addr)
	log.Fatal(srv.ListenAndServe())
}