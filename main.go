package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type conf struct {
	Dependencies []serviceEntry `yaml:"dependencies"`
}

type serviceEntry struct {
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
	Host   string `yaml:"host"`
}

var (
	listenAddr     = flag.String("httpAddr", ":8080", "HTTP address used to listen")
	confFile       = flag.String("conf", "./dependencies.yaml", "Path to configuration file")
	waitTime       = flag.Int("wait", 20, "Wait this time (in ms) before responding the request")
	welcomeMessage = flag.String("welcomeMessage", "Welcome", "Message appended to the response")
)

func main() {

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	s := make(chan os.Signal)
	signal.Notify(s, syscall.SIGTERM, syscall.SIGINT)

	go func() {

		select {
		case <-s:
			logrus.Info("Received a stop signal. Exiting...")
			cancel()

		case <-ctx.Done():

		}

	}()

	conf := readConfig(*confFile)

	server := &http.Server{
		Addr:         *listenAddr,
		Handler:      newRouter(conf, *waitTime),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {

		<-ctx.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)

		if err := server.Shutdown(shutdownCtx); err != nil {
			logrus.Errorf("Could not gracefully shutdown the server: %v\n", err)
		}

		shutdownCancel()

	}()

	logrus.Infof("Starting http server at %s", *listenAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.Errorf("Could not listen on %s: %v\n", *listenAddr, err)
	}

}

func readConfig(path string) *conf {

	c := &conf{}

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c

}

func newRouter(conf *conf, waitTime int) *http.ServeMux {

	router := http.NewServeMux()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		logrus.Infof("%s %s Host: %s", r.Method, r.RequestURI, r.Host)

		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		io.WriteString(w, *welcomeMessage+"\n\n")

		io.WriteString(w, "Headers:\n")

		for key, values := range r.Header {
			io.WriteString(w, fmt.Sprintf("%s: %v\n", key, values))
		}

		io.WriteString(w, "\n")

		for _, svc := range conf.Dependencies {

			req, err := http.NewRequest(svc.Method, svc.Path, nil)

			if err != nil {

				io.WriteString(w, fmt.Sprintf("Dependency: %s Error: %v\n", svc.Path, err))
				continue

			}

			for key, values := range r.Header {
				for _, value := range values {
					req.Header.Add(key, value)
				}
			}

			if len(svc.Host) > 0 {
				req.Host = svc.Host
			}

			resp, err := client.Do(req)

			if err != nil {

				io.WriteString(w, fmt.Sprintf("Dependency: %s Error: %v", req.Host, err))

				continue

			}

			io.WriteString(w, fmt.Sprintf("Dependency: %s Status: %v\n", req.Host, resp.StatusCode))

			io.WriteString(w, "-------\n")

			if body, err := ioutil.ReadAll(resp.Body); err == nil {

				bodyString := "\t" + strings.ReplaceAll(string(body), "\n", "\n\t") + "\n"

				io.WriteString(w, bodyString)

			} else {

				io.WriteString(w, fmt.Sprintf("Cannot read body: %v\n", err))

			}

			resp.Body.Close()

			io.WriteString(w, "-------\n")

			for key, values := range resp.Header {
				for _, value := range values {
					w.Header().Add(key, value)
				}
			}

		}

		for key, values := range r.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		time.Sleep(time.Duration(waitTime) * time.Millisecond)

		hostname, _ := os.Hostname()

		io.WriteString(w, fmt.Sprintf("Processed by %s", hostname))

	})

	return router

}
