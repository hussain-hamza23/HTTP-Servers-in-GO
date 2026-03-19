package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

type filePaths struct {
	app string
	assets string
}

type apiConfig struct{
	fileserverHits atomic.Int32
}

func getHandlers(mux *http.ServeMux, fileDirs filePaths, cfg *apiConfig){
	mux.Handle("/app/", cfg.middlewareMetricsIncrement(http.StripPrefix("/app", http.FileServer(http.Dir(fileDirs.app)))))
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(fileDirs.assets))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", cfg.numberofHits)
	mux.HandleFunc("POST /admin/reset", cfg.resetHits)
}

func handlerReadiness(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := w.Write([]byte("OK")); err != nil {
		http.Error(w, "Status is not OK", http.StatusInternalServerError)
		return
	}

}


func main(){
	const addr string = ":8080"
	var cfg apiConfig = apiConfig{
		fileserverHits: atomic.Int32{},
	}
	var fileDirs filePaths = filePaths{
		app: "./app/",
		assets: "./assets/",
	}

	var mux *http.ServeMux = http.NewServeMux()

	getHandlers(mux, fileDirs, &cfg)

	var srv *http.Server = &http.Server{
		Handler: mux,
		Addr: addr,
	}
	
	log.Printf("Serving files from %s on port %s\n", fileDirs.app, addr)
	log.Fatal(srv.ListenAndServe())
}