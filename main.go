package main

import (
	"drone/droneKit"
	"drone/pkg/api/server"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/url"
	"os"
)

type UTCFormatter struct{}

func (f UTCFormatter) Format(e *logrus.Entry) ([]byte, error) {
	var err error
	var out []byte
	out = []byte(fmt.Sprintf("datatime=%s ,level=%s, msg=%s\n", e.Time.Format("02-01-2006 15:04:05"), e.Level, e.Message))
	return out, err
}

func main() {
	Formatter := new(UTCFormatter)
	logrus.SetFormatter(Formatter)
	logrus.SetLevel(logrus.DebugLevel)

	mode := flag.String("mode", "", "Set app mode")
	flag.Parse()

	token := os.Getenv("tgToken")
	appUrl := os.Getenv("url")

	if appUrl == "" {
		fmt.Println("Wrong url parameter.")
	}
	raw, err := url.Parse(appUrl)
	if err != nil {
		panic(err)
	}

	switch *mode {
	case "client":
		droneKit.Run(raw.Host)

	case "server":
		if token == "" {
			fmt.Println("You need to set token.")
			return
		}
		server.Run(raw.Host, raw.Path, token)

	default:
		fmt.Println("Please enter a valid app mode")
	}
}
