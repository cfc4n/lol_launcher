package replay

import (
	"encoding/json"
	"log"
	"net/http"
	"github.com/gorilla/mux"
)

func FeaturedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(featured())
}

func VersionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.Write([]byte(version()))
}

func GetGameMetaDataHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	data, err := getGameMetaData(vars["platformId"], vars["gameId"])

	j, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(nil)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

func GetLastChunkInfoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chunkInfo, err := getLastChunkInfo(vars["platformId"], vars["gameId"], vars["param"])
	if err != nil {
		log.Println("error", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(nil)
		return
	}

	j, err := json.Marshal(chunkInfo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(nil)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

func EndOfGameStatsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write(nil)
}

func GetGameDataChunkHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	data, err := getGameDataChunk(vars["platformId"], vars["gameId"], vars["chunkId"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(nil)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func GetKeyFrameHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	data, err := getKeyFrame(vars["platformId"], vars["gameId"], vars["keyFrameId"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(nil)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}