package drone

import (
	"drone/pkg/api/server"
	"fmt"
	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
	"sync"
)

type Drone struct {
	channel    *gomavlib.Channel
	client     *gomavlib.Node
	GpsUpdates *RingBuffer
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

	dr := Drone{
		channel:    d,
		client:     node,
		GpsUpdates: NewRingBuffer(8),
	}

	go dr.gpsUpdater()
	return &dr, nil
}

type RingBuffer struct {
	buffer []server.Message
	size   int
	head   int
	tail   int
	mutex  sync.Mutex
}

// Создание нового кольцевого буфера
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		buffer: make([]server.Message, size),
		size:   size,
		head:   0,
		tail:   0,
	}
}

// Добавление элемента в буфер
func (rb *RingBuffer) Push(msg server.Message) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	// Проверка на переполнение
	if (rb.tail+1)%rb.size == rb.head {
		rb.head = (rb.head + 1) % rb.size // Перемещаем голову, чтобы освободить место
	}

	rb.buffer[rb.tail] = msg
	rb.tail = (rb.tail + 1) % rb.size
}

// Извлечение элемента из буфера
func (rb *RingBuffer) Pop() (server.Message, bool) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	// Проверка на пустоту
	if rb.head == rb.tail {
		return server.Message{}, false
	}

	msg := rb.buffer[rb.head]
	rb.head = (rb.head + 1) % rb.size
	return msg, true
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
				d.GpsUpdates.Push(server.Message{
					Lat:     float64(msg.Lat) / 10000000,
					Lon:     float64(msg.Lon) / 10000000,
					Alt:     float64(msg.RelativeAlt) / 1000,
					Heading: float64(msg.Hdg) / 100,
				})
			}
		}
	}
}
