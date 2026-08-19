package main

import (
	"flag"
)

// flagHTTPAddr holds the HTTP listen address.
var flagHTTPAddr string

// flagLogLevel holds the server log level.
var flagLogLevel string

// TODO
var flagStoreInterval int

// TODO
var flagFileStoragePath string

// TODO
var flagRestore bool

// parseFlags registers and parses command-line flags.
func parseFlags() {
	flag.StringVar(&flagHTTPAddr, "a", "localhost:8080", "HTTP server endpoint address")
	flag.StringVar(&flagLogLevel, "l", "Info", "HTTP server log level")
	flag.IntVar(&flagStoreInterval, "i", 300, "интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск. Значение 0 делает запись синхронной ")
	flag.StringVar(&flagFileStoragePath, "f", "storage.json", "путь до файла, куда сохраняются текущие значения. Имя файла для значения по умолчанию придумайте сами")
	flag.BoolVar(&flagRestore, "r", true, "следует ли загружать ранее сохранённые значения из указанного файла при старте сервера")
	flag.Parse()
}
