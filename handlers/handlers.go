package handlers

import (
	"lightsites/config"
	"lightsites/document"

	"fmt"
	"log"
	"net/http"
	"strings"
)

func ContentHandler(w http.ResponseWriter, req *http.Request, documents *[]document.Document, conf *config.Config) {
	// w.Header().Set("Access-Control-Allow-Origin", "*")
	// w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	// w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)

	// handle preflight options requests
	if req.Method == "OPTIONS" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		return
	}

	documentName := strings.TrimPrefix(req.URL.Path, fmt.Sprintf("%v", conf.Routing.RoutePrefix))

	if documentName == "" {
		documentName = fmt.Sprintf("index%v", conf.Routing.UrlFileSuffix)
	}

	for _, document := range *documents {
		if fmt.Sprintf("%v%v", document.DocumentName, conf.Routing.UrlFileSuffix) == documentName {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			result, err := w.Write([]byte(document.FileContents))
			if err != nil {
				log.Printf("failed to write http response: %v", err.Error())
			}
			log.Printf("%v transferred %v bytes", req.URL.Path, result)
			return
		}
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusNotFound)
	result, err := w.Write([]byte{})
	if err != nil {
		log.Printf("failed to write http response: %v", err.Error())
	}
	log.Printf("%v transferred %v bytes", req.URL.Path, result)
}
