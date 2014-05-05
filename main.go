package main

import (
	"flag"
	"github.com/mvader/sunglasses/app"
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
		if err := http.ListenAndServeTLS(a.Config.Port, a.Config.SSLCert, a.Config.SSLKey, a.Martini); err != nil {
			a.Logger.Fatal(err)
		}
	} else {
		if err := http.ListenAndServe(a.Config.Port, a.Martini); err != nil {
			a.Logger.Fatal(err)
		}
	}
}
