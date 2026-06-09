package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"scoreador/internal/store"
	"scoreador/internal/webapp"
)

func main() {
	var (
		addr       = flag.String("addr", ":8081", "Direccion del servidor web")
		dbPath     = flag.String("db", filepath.Join("out", "single_matches.db"), "Ruta de la base SQLite")
		lambdaPath = flag.String("lambda", "examples/demo/lambda.csv", "Ruta por defecto de la tabla lambda")
	)
	flag.Parse()

	if err := os.MkdirAll(filepath.Dir(*dbPath), 0o755); err != nil {
		fatal(err)
	}

	db, err := store.OpenSQLite(*dbPath)
	if err != nil {
		fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error cerrando sqlite: %v", err)
		}
	}()

	app := webapp.NewServer(db, *lambdaPath, filepath.Join("out", "web_batch"))
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		fatal(err)
	}
	fmt.Printf("GUI lista en http://localhost%s\n", *addr)
	fmt.Printf("Base SQLite: %s\n", *dbPath)
	if err := http.Serve(ln, app.Routes()); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	log.Fatal(err)
}
