package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/yodstar/goutil/logger"

	"tunproxy/config"
	"tunproxy/startup"
)

var (
	CONF = config.CONF
	LOG  = logger.LOG
)

func main() {
	var command *string
	if runtime.GOOS == "windows" {
		command = flag.String("s", "", "install, uninstall, start, stop")
	}
	logpath := flag.String("o", "./tunproxy.out", "Stdlog file path")
	cfgpath := flag.String("c", "./tunproxy.conf", "Config file path")
	flag.Parse()

	if exefile, err := os.Executable(); err != nil {
		panic(err.Error())
	} else if err = os.Chdir(filepath.Dir(exefile)); err != nil {
		panic(err.Error())
	}

	logfile, err := os.OpenFile(*logpath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		panic(err.Error())
	}

	log.SetOutput(logfile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	if err = config.LoadFile(*cfgpath); err != nil {
		log.Println(err)
		return
	}

	// logger
	LOG.SetLevel(CONF.Logger.Level)
	LOG.SetOutFile(CONF.Logger.Outfile, "200601")
	LOG.SetFilter(CONF.Logger.Filter, func(s string) { log.Output(4, s) })
	LOG.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	startup.Run(command)
}
