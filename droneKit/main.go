package droneKit

import (
	"drone/pkg/api/client"
	"drone/pkg/drone"
	"drone/pkg/modem"
	"time"
)

type DroneKit struct {
	Drone    *drone.Drone
	Modem    *modem.Modem
	BlackBox *client.Client
}

var droneKit *DroneKit

func Run(domain string) {

	dk := new(DroneKit)
	var err error
	dk.Drone, err = drone.New()
	if err != nil {
		panic(err)
	}
	dk.Modem, err = modem.New()
	if err != nil {
		panic(err)
	}
	droneKit = dk

	droneKit.BlackBox = client.New(domain)
	go droneKit.sendLinkState()
	droneKit.sendGPS()
}

func (kit *DroneKit) sendLinkState() {
	for {
		ls := <-kit.Modem.LinkState
		kit.Drone.SendText(ls.String())
	}
}

func (kit *DroneKit) sendGPS() {
	for {
		msg, err := kit.Drone.GpsUpdates.Pop()
		if err == false {
			time.Sleep(1 * time.Second)
			continue
		}

		kit.BlackBox.SendWithTimeout(msg)
		time.Sleep(1 * time.Second)
	}
}
