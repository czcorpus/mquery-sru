// Copyright 2019 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2019 Institute of the Czech National Corpus,
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

package cnf

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/bytedance/sonic"
	"github.com/czcorpus/mquery-sru/corpus"
	"github.com/czcorpus/mquery-sru/rdb"

	"github.com/czcorpus/cnc-gokit/logging"
	"github.com/rs/zerolog/log"
)

const (
	dfltServerWriteTimeoutSecs = 30
	dfltLanguage               = "en"
	dfltMaxNumConcurrentJobs   = 4
	dfltVertMaxNumErrors       = 100

	dfltTimeZone       = "Europe/Prague"
	dfltSourcesRootDir = "."
	dfltAssetsURLPath  = "/"
)

type ServerInfo struct {
	ServerName string `json:"serverName"`
	ServerPort string `json:"serverPort"`
	Database   string `json:"database"`

	// language mappings
	DatabaseTitle       map[string]string `json:"databaseTitle"`       // section required, "en" required
	DatabaseDescription map[string]string `json:"databaseDescription"` // section optional, "en" required
	PrimaryLanguage     string            `json:"primaryLanguage"`
}

func (s *ServerInfo) Validate() error {
	if s == nil {
		return errors.New("missing serverInfo section")
	}
	if s.DatabaseTitle == nil {
		return errors.New("missing configuration section `serverInfo.databaseTitle`")
	}
	_, ok := s.DatabaseTitle["en"]
	if !ok {
		return errors.New("missing required configuration for `serverInfo.databaseTitle.en`")
	}

	if s.DatabaseDescription != nil {
		_, ok := s.DatabaseDescription["en"]
		if !ok {
			return errors.New("missing required configuration for `serverInfo.databaseDescription.en`")
		}
	}

	return nil
}

// Conf is a global configuration of the app
type Conf struct {
	ListenAddress          string   `json:"listenAddress"`
	ListenPort             int      `json:"listenPort"`
	ServerReadTimeoutSecs  int      `json:"serverReadTimeoutSecs"`
	ServerWriteTimeoutSecs int      `json:"serverWriteTimeoutSecs"`
	CorsAllowedOrigins     []string `json:"corsAllowedOrigins"`

	// SourcesRootDir is mainly used to locate html/xml templates and other
	// assets so we can refer them in a relative way inside the code
	SourcesRootDir string               `json:"sourcesRootDir"`
	AssetsURLPath  string               `json:"assetsURLPath"`
	ServerInfo     *ServerInfo          `json:"serverInfo"`
	CorporaSetup   *corpus.CorporaSetup `json:"corpora"`
	Redis          *rdb.Conf            `json:"redis"`
	LogFile        string               `json:"logFile"`
	LogLevel       logging.LogLevel     `json:"logLevel"`
	Language       string               `json:"language"`
	TimeZone       string               `json:"timeZone"`

	srcPath string
}

func (conf *Conf) IsDebugMode() bool {
	return conf.LogLevel == "debug"
}

func (conf *Conf) TimezoneLocation() *time.Location {
	// we can ignore the error here as we always call c.Validate()
	// first (which also tries to load the location and report possible
	// error)
	loc, _ := time.LoadLocation(conf.TimeZone)
	return loc
}

// GetSourcePath returns an absolute path of a file
// the config was loaded from.
func (conf *Conf) GetSourcePath() string {
	if filepath.IsAbs(conf.srcPath) {
		return conf.srcPath
	}
	var cwd string
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "[failed to get working dir]"
	}
	return filepath.Join(cwd, conf.srcPath)
}

func LoadConfig(path string) *Conf {
	if path == "" {
		log.Fatal().Msg("Cannot load config - path not specified")
	}
	rawData, err := os.ReadFile(path)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot load config")
	}
	var conf Conf
	conf.srcPath = path
	err = sonic.Unmarshal(rawData, &conf)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot load config")
	}
	return &conf
}

func ValidateAndDefaults(conf *Conf) {
	if conf.ServerWriteTimeoutSecs == 0 {
		conf.ServerWriteTimeoutSecs = dfltServerWriteTimeoutSecs
		log.Warn().Msgf(
			"serverWriteTimeoutSecs not specified, using default: %d",
			dfltServerWriteTimeoutSecs,
		)
	}
	if err := conf.ServerInfo.Validate(); err != nil {
		log.Fatal().Err(err).Msg("invalid configuration")
		return
	}
	if conf.Language == "" {
		conf.Language = dfltLanguage
		log.Warn().Msgf("language not specified, using default: %s", conf.Language)
	}
	if err := conf.CorporaSetup.ValidateAndDefaults("corpora"); err != nil {
		log.Fatal().Err(err).Msg("invalid configuration")
		return
	}
	if err := conf.CorporaSetup.ValidateAndDefaults("corporaSetup"); err != nil {
		log.Fatal().Err(err).Msg("invalid configuration")
		return
	}
	if conf.TimeZone == "" {
		log.Warn().
			Str("timeZone", dfltTimeZone).
			Msg("time zone not specified, using default")
	}
	if _, err := time.LoadLocation(conf.TimeZone); err != nil {
		log.Fatal().Err(err).Msg("invalid time zone")
		return
	}
	if conf.SourcesRootDir == "" {
		log.Warn().
			Str("sourcesRootDir", dfltSourcesRootDir).
			Msg("sources root directory not specified, using default")
		conf.SourcesRootDir = dfltSourcesRootDir
	}
	if conf.AssetsURLPath == "" {
		log.Warn().
			Str("assetsURLPath", dfltAssetsURLPath).
			Msg("URL path of assets not set, using default (this is needed only for UI features)")
		conf.AssetsURLPath = dfltAssetsURLPath
	}
}
