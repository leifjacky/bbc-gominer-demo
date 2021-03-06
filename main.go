package main

import (
	"io"
	"os"
	"runtime"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type StratumMinerConfig struct {
	Url      string
	Username string
	Password string
	SumIntv  string
	Threads  int
	UseTls   bool
	PemFile  string
	KeyFile  string
}

func main() {
	var url, username, password, loglevel, logfile, pemfile, keyfile string
	var useTls bool
	var threads int
	pflag.StringVarP(&url, "url", "o", "bbc.uupool.cn:12812", "stratum pool url")
	pflag.StringVarP(&username, "username", "u", "1rmycpneemett21ejf6vnan0twktdqah5scqd2ybrkhfyf2x9yr038r44.worker1", "username")
	pflag.StringVarP(&password, "password", "x", "x", "password")
	pflag.StringVarP(&loglevel, "loglevel", "l", "debug", "log level: info, debug, trace")
	pflag.StringVarP(&logfile, "logfile", "f", "debug.log", "logfile path")
	pflag.IntVarP(&threads, "threads", "t", runtime.NumCPU(), "threads")
	pflag.BoolVarP(&useTls, "tls", "s", false, "use tls")
	pflag.StringVarP(&pemfile, "pemfile", "p", "tls.pem", "pemfile path for tls")
	pflag.StringVarP(&keyfile, "keyfile", "k", "tls.key", "keyfile path for tls")
	pflag.Parse()

	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
	if l, err := logrus.ParseLevel(loglevel); err != nil {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(l)
	}

	if logfile == "" {
		logrus.Warningf("Ignore logging to file")
	}
	ljack := &lumberjack.Logger{
		Filename: logfile,
	}
	mWriter := io.MultiWriter(os.Stdout, ljack)
	logrus.SetOutput(mWriter)

	cfg := &StratumMinerConfig{
		Url:      url,
		Username: username,
		Password: password,
		SumIntv:  "10s",
		Threads:  threads,
		UseTls:   useTls,
		PemFile:  pemfile,
		KeyFile:  keyfile,
	}
	m := NewMiner(cfg)
	m.Mine()
}
