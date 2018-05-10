package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"

	"github.com/ProdriveTechnologies/snippets/pkg/util"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var (
		dbAddress   = flag.String("db.address", "", "Database server address.")
		snippetsUrl = flag.String("snippets.url", "", "URL of the Snippets site.")
	)
	flag.Parse()

	db, err := gorm.Open("postgres", *dbAddress)
	if err != nil {
		panic(err)
	}

	templates, err := template.ParseGlob("templates/*")
	if err != nil {
		panic(err)
	}

	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())
	util.RegisterHealthPage(db, router)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	NewSnippetsWebService(db, templates, *snippetsUrl, router)
	log.Fatal(http.ListenAndServe(":80", router))
}
