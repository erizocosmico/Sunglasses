package main

import (
	"flag"
	"github.com/mvader/mask/app"
	"net/http"
	"os"
)

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	configPath := fs.String("config", "config.json", "Path to config.json")
	fs.Parse(os.Args[1:])

	a, err := app.NewApp(*configPath)
	if err != nil {
		panic(err)
	}

	defer func() {
		a.Connection.Session.Close()
		a.LogFile.Close()
	}()

	if a.Config.UseHTTPS {
		// TODO
	} else {
		http.ListenAndServe(a.Config.Port, a.Martini)
	}
}
