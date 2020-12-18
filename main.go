package main

import (
	"lightsites/config"
	"lightsites/document"
	"lightsites/handlers"
	"lightsites/helpers"

	"fmt"
	"log"
	"net/http"
	"time"
)

var documents []document.Document
var documentDirectoryList helpers.DirectoryListing
var globalConf *config.Config

func contentHandler(w http.ResponseWriter, req *http.Request) {
	handlers.ContentHandler(w, req, &documents, globalConf)
}

func main() {
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to process config: %v", err.Error())
	}

	globalConf = &conf

	go func() {
		for {
			log.Print("reading directory...")
			documentDirectoryList.Path = conf.Directories.Documents

			err := documentDirectoryList.WalkDirectory()
			if err != nil {
				log.Fatalf("failed to read directory %v: %v", conf.Directories.Documents, err.Error())
			}

			documents = []document.Document{}
			for _, file := range documentDirectoryList.Files {
				_, err := document.ParseDocument(&conf, &documents, &documentDirectoryList.Files, file)
				if err != nil {
					log.Printf("failed to process document %v: %v", file, err.Error())
				}
			}
			log.Printf("done reading directory. %v documents found. sleeping %v.", len(documents), conf.RefreshInterval)
			time.Sleep(conf.RefreshInterval)
		}
	}()

	http.HandleFunc(fmt.Sprintf("%v", conf.Routing.RoutePrefix), contentHandler)

	// serve static files
	fs := http.FileServer(http.Dir(conf.Directories.Assets))
	http.Handle(conf.Routing.AssetsPrefix, http.StripPrefix(conf.Routing.AssetsPrefix, fs))

	log.Printf("begin listening on %v", conf.ListenAddr)
	err = http.ListenAndServe(conf.ListenAddr, nil)
	if err != nil {
		log.Fatalf("failed to listen and serve: %v", err.Error())
	}
}
