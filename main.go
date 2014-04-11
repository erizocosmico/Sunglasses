package main

import (
	"flag"
	. "github.com/mvader/mask/mask"
	"net/http"
	"os"
	"github.com/mvader/mask/app"
)

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	configPath := fs.String("config", "config.json", "Path to config.json")
	fs.Parse(os.Args[1:])

	app, port, err := app.NewApp(*configPath)
	if err != nil {
		panic(err)
	}

	http.ListenAndServe(port, app)
}
