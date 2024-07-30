package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	tb "gopkg.in/telebot.v3"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"time"
)

type Message struct {
	Time    int64
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Alt     float64 `json:"alt"`
	Heading float64 `json:"heading"`
}

func (msg Message) String() string {
	if msg.Time == 0 {
		msg.Time = time.Now().Unix()
	}
	t := time.Unix(msg.Time, 0)
	return fmt.Sprintf("%s alt: %.2f, heading: %.2f, https://www.google.com/maps/place/%.8f,%.8f", t.Format(time.DateTime), msg.Alt, msg.Heading, msg.Lat, msg.Lon)
}

var db *sql.DB
var tg *tb.Bot

func Run(domain, endpoint, tgToken string) {

	defer func() {
		if r := recover(); r != nil {
			logrus.Debug(r)
			logrus.Debug(string(debug.Stack()))
		}
	}()

	_, err := os.Stat("./db/droneGPS.db")
	if os.IsNotExist(err) {
		file, err := os.Create("droneGPS.db")
		if err != nil {
			logrus.Fatal(err.Error())
		}
		file.Close()
	}

	db, _ = sql.Open("sqlite3", "./db/droneGPS.db")
	err = createTable()
	if err != nil {
		logrus.Fatal(err.Error())
	}

	hook := &tb.Webhook{
		Endpoint: &tb.WebhookEndpoint{PublicURL: fmt.Sprintf("https://%s%s", domain, endpoint)},
	}
	s := tb.Settings{
		Token:  tgToken,
		Poller: hook,
	}

	logrus.Debugf("Create tg bot, with webhook on https://%s%s", domain, endpoint)
	tg, err = tb.NewBot(s)
	if err != nil {
		logrus.Debug(err.Error())
	}
	botHandlers()
	go tg.Start()

	mux := http.NewServeMux()
	mux.Handle("/put", putHandler())
	mux.Handle(endpoint, hook)
	logrus.Debugf("Listening on %s:%s", domain, "8094")
	logrus.Debug(http.ListenAndServe(":8094", mux))
}

func createTable() error {
	query := `CREATE TABLE IF NOT EXISTS gps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    unixtime INTEGER NOT NULL,
    lat REAL NOT NULL,
    lon REAL NOT NULL,
    alt REAL NOT NULL,
    heading REAL NOT NULL
	);`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func writeRecord(msg *Message) error {
	query := "INSERT INTO gps (unixtime, lat, lon, alt, heading) VALUES (?, ?, ?, ?, ?)"
	_, err := db.Exec(query, time.Now().Unix(), msg.Lat, msg.Lon, msg.Alt, msg.Heading)
	return err
}

func botHandlers() {
	cId, _ := strconv.Atoi(os.Getenv("chatid"))

	tg.Handle(tb.OnText, func(c tb.Context) error {
		if c.Message().Chat.ID != int64(cId) {
			return nil
		}
		msgs := readRecord()
		out := ""
		for _, msg := range msgs {
			out += msg.String() + "\n"
		}

		_, err := tg.Send(c.Message().Chat, out)
		if err != nil {
			logrus.Debug(err.Error())
		}
		return nil
	})

}

func readRecord() []Message {
	query := "SELECT unixtime ,lat, lon, alt, heading FROM gps order by id desc limit 30"
	rows, err := db.Query(query)
	if err != nil {
		logrus.Debug(err.Error())
		return nil
	}
	var msgs []Message

	for rows.Next() {
		var msg Message
		err = rows.Scan(&msg.Time, &msg.Lat, &msg.Lon, &msg.Alt, &msg.Heading)
		if err != nil {
			logrus.Debug(err.Error())
			continue
		}
		msgs = append(msgs, msg)
	}

	return msgs
}

func putHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logrus.Debug(r)
				logrus.Debug(string(debug.Stack()))
			}
			w.WriteHeader(200)
		}()

		raw, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(500)
		}

		msg := new(Message)
		err = json.Unmarshal(raw, msg)
		if err != nil {
			logrus.Debug(fmt.Errorf("Error unmarshalling message: %v", err))
		}

		err = writeRecord(msg)
		if err != nil {
			logrus.Debug(fmt.Errorf("Error writing message: %v", err))
		}
	})
}
