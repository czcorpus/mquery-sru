// Copyright 2023 Martin Zimandl <martin.zimandl@gmail.com>
// Copyright 2024 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2023 Institute of the Czech National Corpus,
//                Faculty of Arts, Charles University
//   This file is part of MQUERY.
//
//  MQUERY is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  MQUERY is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with MQUERY.  If not, see <https://www.gnu.org/licenses/>.

//go:generate pigeon -o query/parser/fcsql/fcsql.go query/parser/fcsql/fcsql.peg
//go:generate pigeon -o query/parser/simple/simple.go query/parser/simple/simple.peg

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/czcorpus/cnc-gokit/logging"
	"github.com/czcorpus/cnc-gokit/uniresp"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"fcs/cnf"
	"fcs/general"
	"fcs/handler"
	"fcs/handler/form"
	"fcs/monitoring"
	"fcs/rdb"
	"fcs/worker"
)

var (
	version   string
	buildDate string
	gitCommit string
)

func getEnv(name string) string {
	for _, p := range os.Environ() {
		items := strings.Split(p, "=")
		if len(items) == 2 && items[0] == name {
			return items[1]
		}
	}
	return ""
}

func init() {
}

func runApiServer(
	conf *cnf.Conf,
	syscallChan chan os.Signal,
	exitEvent chan os.Signal,
	radapter *rdb.Adapter,
) {
	if !conf.LogLevel.IsDebugMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(logging.GinMiddleware())
	engine.NoMethod(uniresp.NoMethodHandler)
	engine.NoRoute(uniresp.NotFoundHandler)

	FCSActions := handler.NewFCSHandler(
		conf.ServerInfo, conf.CorporaSetup, radapter, conf.SourcesRootDir)
	engine.GET("/", FCSActions.FCSHandler)
	engine.HEAD("/", FCSActions.FCSHandler)

	uIActions := form.NewFormHandler(conf.CorporaSetup, conf.SourcesRootDir)
	engine.GET("/ui/form", uIActions.Handle)

	logger := monitoring.NewWorkerJobLogger(conf.TimezoneLocation())
	logger.GoRunTimelineWriter()

	monitoringActions := monitoring.NewActions(logger, conf.TimezoneLocation())
	engine.GET("/monitoring/workers-load", monitoringActions.WorkersLoad)

	log.Info().Msgf("starting to listen at %s:%d", conf.ListenAddress, conf.ListenPort)
	srv := &http.Server{
		Handler:      engine,
		Addr:         fmt.Sprintf("%s:%d", conf.ListenAddress, conf.ListenPort),
		WriteTimeout: time.Duration(conf.ServerWriteTimeoutSecs) * time.Second,
		ReadTimeout:  time.Duration(conf.ServerReadTimeoutSecs) * time.Second,
	}
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Error().Err(err).Msg("")
		}
		syscallChan <- syscall.SIGTERM
	}()

	select {
	case <-exitEvent:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := srv.Shutdown(ctx)
		if err != nil {
			log.Info().Err(err).Msg("Shutdown request error")
		}
	}
}

func runWorker(conf *cnf.Conf, workerID string, radapter *rdb.Adapter, exitEvent chan os.Signal) {
	ch := radapter.Subscribe()
	logger := monitoring.NewWorkerJobLogger(conf.TimezoneLocation())
	w := worker.NewWorker(workerID, radapter, ch, exitEvent, logger)
	w.Listen()
}

func getWorkerID() (workerID string) {
	workerID = getEnv("WORKER_ID")
	if workerID == "" {
		workerID = "0"
	}
	return
}

func main() {
	version := general.VersionInfo{
		Version:   version,
		BuildDate: buildDate,
		GitCommit: gitCommit,
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "CNC-FCS - A specialized corpus querying server\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options] server [config.json]\n\t", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options] worker [config.json]\n\t", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s transform [basic/advanced]\n\t", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "%s [options] version\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()
	action := flag.Arg(0)
	switch action {
	case "version":
		fmt.Printf("cnc-fcs %s\nbuild date: %s\nlast commit: %s\n", version.Version, version.BuildDate, version.GitCommit)
		return
	case "transform":
		switch flag.Arg(1) {
		case "basic":
			repl(translateBasicQuery)
		case "advanced":
			repl(translateFCSQuery)
		default:
			fmt.Println("Unknown query type")
			os.Exit(2)
		}
	}

	conf := cnf.LoadConfig(flag.Arg(1))

	if action == "worker" {
		var wPath string
		if conf.LogFile != "" {
			wPath = filepath.Join(filepath.Dir(conf.LogFile), "worker.log")
		}
		logging.SetupLogging(wPath, conf.LogLevel)
		log.Logger = log.Logger.With().Str("worker", getWorkerID()).Logger()

	} else if action == "test" {
		cnf.ValidateAndDefaults(conf)
		log.Info().Msg("config OK")
		return

	} else {
		logging.SetupLogging(conf.LogFile, conf.LogLevel)
	}
	log.Info().Msg("Starting CNC-FCS")
	cnf.ValidateAndDefaults(conf)
	syscallChan := make(chan os.Signal, 1)
	signal.Notify(syscallChan, os.Interrupt)
	signal.Notify(syscallChan, syscall.SIGTERM)
	exitEvent := make(chan os.Signal)
	testConnCancel := make(chan bool)
	go func() {
		evt := <-syscallChan
		testConnCancel <- true
		close(testConnCancel)
		exitEvent <- evt
		close(exitEvent)
	}()

	radapter := rdb.NewAdapter(conf.Redis)

	switch action {
	case "server":
		err := radapter.TestConnection(20*time.Second, testConnCancel)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to Redis")
		}
		runApiServer(conf, syscallChan, exitEvent, radapter)
	case "worker":
		err := radapter.TestConnection(20*time.Second, testConnCancel)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to Redis")
		}
		runWorker(conf, getWorkerID(), radapter, exitEvent)
	default:
		log.Fatal().Msgf("Unknown action %s", action)
	}

}
