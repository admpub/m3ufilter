package server

import (
	"github.com/hoshsadiq/m3ufilter/config"
	"github.com/hoshsadiq/m3ufilter/logger"
	"github.com/hoshsadiq/m3ufilter/m3u"
	"github.com/hoshsadiq/m3ufilter/writer"
	"github.com/mileusna/crontab"
	"net/http"
)

var log = logger.Get()

var playlists *m3u.Streams

var lock bool

func Serve(conf *config.Config) {
	schedule := conf.Core.UpdateSchedule
	if schedule == "" {
		schedule = "*/24 * * * *"
	}

	if playlists == nil {
		playlists = &m3u.Streams{}
	}

	log.Info("Scheduling cronjob to periodically update playlist.")
	ctab := crontab.New()
	ctab.MustAddJob(conf.Core.UpdateSchedule, func() {
		updatePlaylist(conf)
	})

	log.Info("Parsing for the first time...")
	ctab.RunAll()

	log.Info("starting server")
	http.HandleFunc("/playlist.m3u", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "audio/mpegurl")

		writer.WriteOutput(conf.Core.Output, w, *playlists)
	})

	server := &http.Server{Addr: conf.Core.ServerListen}
	log.Fatal(server.ListenAndServe())
}

func updatePlaylist(conf *config.Config) {
	if lock {
		log.Info("Retrieval is locked, trying again next time...")
		return
	}

	lock = true
	log.Info("updating playlists")
	newPlaylists := m3u.GetPlaylist(conf)
	playlists = &newPlaylists
	log.Info("done")
	lock = false
}
