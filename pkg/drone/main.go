package drone

import (
	"drone/pkg/api/server"
	"fmt"
	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
)

type Drone struct {
	channel    *gomavlib.Channel
	client     *gomavlib.Node
	GpsUpdates chan server.Message
}

func New() (*Drone, error) {
	node, err := gomavlib.NewNode(gomavlib.NodeConf{
		Endpoints: []gomavlib.EndpointConf{
			gomavlib.EndpointUDPClient{
				Address: "127.0.0.1:5600",
			},
		},
		Dialect:     ardupilotmega.Dialect,
		OutVersion:  gomavlib.V2,
		OutSystemID: 10,
	})
	if err != nil {
		return nil, err
	}

	var d *gomavlib.Channel

	for evt := range node.Events() {
		if frm, ok := evt.(*gomavlib.EventFrame); ok {
			//
			if int(frm.ComponentID()) == 1 && int(frm.SystemID()) == 1 {
				d = frm.Channel
				break
			}
		}
	}

	gps := make(chan server.Message)
	dr := Drone{
		channel:    d,
		client:     node,
		GpsUpdates: gps,
	}

	go dr.gpsUpdater()
	return &dr, nil
}

func (d *Drone) SendText(text string) {
	d.client.WriteMessageTo(d.channel, &ardupilotmega.MessageStatustext{
		Text:     fmt.Sprintf(text),
		Severity: ardupilotmega.MAV_SEVERITY_INFO,
	})
}

func (d *Drone) gpsUpdater() {
	for evt := range d.client.Events() {
		if frm, ok := evt.(*gomavlib.EventFrame); ok {
			switch msg := frm.Message().(type) {

			case *ardupilotmega.MessageGlobalPositionInt:
				//if msg.Lon == 0 {
				//	continue
				//}
				d.GpsUpdates <- server.Message{
					Lat:     float64(msg.Lat) / 10000000,
					Lon:     float64(msg.Lon) / 10000000,
					Alt:     float64(msg.RelativeAlt) / 1000,
					Heading: float64(msg.Hdg) / 100,
				}
			}
		}
	}
}
