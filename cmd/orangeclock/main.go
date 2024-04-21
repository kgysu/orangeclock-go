package main

import (
	"fmt"
	"log"
	"log/slog"
	"machine"
	"orangeclock/pkg/epd2in9v2"
	"orangeclock/pkg/http"
	"strings"
	"time"
)

const displayFullReloadInterval = 12 * time.Hour
const requestDataInterval = 5 * time.Minute
const targetDataServerAddr = "192.168.1.55:48080"
const targetRequestPath = "/mempool/api/orangeclock"

func main() {
	for {
		if err := run(); err != nil {
			panic(err)
		}
		log.Println("restart in 5s!")
		time.Sleep(5 * time.Second)
	}
}

func run() error {
	time.Sleep(1 * time.Second)
	logger := slog.New(slog.NewTextHandler(machine.Serial, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}))
	logger.Debug("starting..")

	display := epd2in9v2.NewPaperDisplay(logger)
	httpClient, ssid, err := http.NewHttpClient(logger, targetDataServerAddr)
	if err != nil {
		return err
	}
	logger.Debug("connected to", slog.String("ssid", ssid))

	// Start
	t := time.Now().Add(displayFullReloadInterval)
	display.UpdateWlanStatus(fmt.Sprintf("#%s", strings.ToUpper(ssid)))
	for {
		logger.Debug("run update cycle")
		if time.Now().After(t) {
			logger.Debug("do a full display reload")
			display.ClearAndSleep()
			t = time.Now().Add(displayFullReloadInterval)
		}

		res := httpClient.NewRequest(targetRequestPath)
		logger.Debug("response to display", slog.String("content", res))
		err := drawLines(display, res)
		if err != nil {
			logger.Error(err.Error())
		}
		time.Sleep(requestDataInterval)
	}

	return nil
}

func drawLines(d *epd2in9v2.PaperDisplay, text string) error {
	lines := strings.Split(text, "\n")
	if len(lines) <= 15 {
		return fmt.Errorf("invalid data input, got lines=%d", len(lines))
	}

	d.UpdateLine(lines[5], 0)
	d.UpdateLineMedium(lines[6], 12)
	d.UpdateLine(lines[7], 30)

	d.UpdateLineMedium(lines[9], 47)

	d.UpdateLine(lines[11], 70)
	d.UpdateLineMedium(lines[12], 82)

	d.UpdateLine(lines[14], 100)
	d.UpdateLineMedium(lines[15], 112)
	return nil
}

func testDrawing(logger *slog.Logger) {
	d := epd2in9v2.NewPaperDisplay(logger)
	text := "HTTP/1.1 200 OK\n" +
		"Date: Sat, 20 Apr 2024 13:42:50 GMT\n" +
		"Content-Length: 128\n" +
		"Content-Type: text/plain; charset=utf-8\n" +
		"\n" +
		"PRICE (15:40:04)\n" +
		" $64,012   +58,213\n" +
		"1$: 1,562\n" +
		"\n" +
		"      840,076 @\n" +
		"\n" +
		"HALVING\n" +
		" 209,924/1,050,000(0%)\n" +
		"\n" +
		"FEES\n" +
		" 541 - 700 - 800\n" +
		"\n" +
		"\n"
	drawLines(d, text)

	time.Sleep(60 * time.Second)
}
