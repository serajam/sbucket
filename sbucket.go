package main

import (
	"flag"
	"fmt"
	"github.com/serajam/sbucket/internal/server"
	"github.com/serajam/sbucket/internal/storage"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()


	storageType := flag.String("type", storage.SyncMapType,
		fmt.Sprintf("Available storage types: %s, %s", storage.MapType, storage.SyncMapType))
	address := flag.String("address", "localhost:3456", "Listening address for server")
	loggerType := flag.String("logger", "", "Logger to use. Default set to go's log. Type `logrus` can be used instead")
	infoLog := flag.String("info_log", "/dev/stdout", "Path for info logging")
	errorLog := flag.String("error_log", "/dev/stderr", "Path for error logging")
	login := flag.String("login", "", "Login for authentication. Used in pair with password")
	password := flag.String("password", "", "Password for authentication. Used in pair with login")
	maxConnNum := flag.Int("max_conn_num", 0, "Maximum number for allowed simulations connections")
	connTimeout := flag.Int("connect_timeout", 5, "Maximum number for allowed simulations connections")

	flag.Parse()

	strg, err := storage.New(*storageType)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	middleware := make([]server.Middleware, 0)
	if len(*login) > 0 && len(*password) > 0 {
		log.Println("Using auth middleware")
		middleware = append(middleware, server.AuthMiddleware{Login: *login, Pass: *password})
	}

	s := server.New(strg,
		server.Address(*address),
		server.Logger(*loggerType, *infoLog, *errorLog),
		server.WithMiddleware(middleware...),
		server.MaxConnNum(*maxConnNum),
		server.ConnectTimeout(*connTimeout),
	)
	s.Start()
}
