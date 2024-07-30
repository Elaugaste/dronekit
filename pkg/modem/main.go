package modem

import (
	"github.com/warthog618/modem/at"
	"os"
	"time"
)

type Modem struct {
	LinkState chan LinkState
}

var client *at.AT
var lastLinkState LinkState

func New() (*Modem, error) {

	f, err := os.OpenFile("/dev/ttyUSB3", os.O_RDWR, os.ModeDevice)
	if err != nil {
		return nil, err
	}
	client = at.New(f)
	client.Escape()

	sig := make(chan LinkState)
	go linkStateUpdater(sig, time.Second*10)

	return &Modem{
		LinkState: sig,
	}, nil
}

func deleteSMS() {
	client.Command("CMGD=0,4")
}

func linkStateUpdater(ls chan LinkState, interval time.Duration) {
	for {
		lq := getLQ()
		time.Sleep(time.Second)
		newState := LinkState{
			Mode:     getMode(),
			Strength: lq[0],
			LQ:       lq[1],
		}

		if newState.Mode != lastLinkState.Mode ||
			newState.Strength != lastLinkState.Strength ||
			newState.LQ != lastLinkState.LQ {

			lastLinkState = newState
			ls <- newState
		}
		time.Sleep(interval)
	}
}
