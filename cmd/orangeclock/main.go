package main

import (
  "errors"
  "fmt"
  "log"
  "log/slog"
  "machine"
  "orangeclock/pkg/epd2in9v2"
  "orangeclock/pkg/http"
  "strings"
  "time"
)

const displayFullReloadInterval = 20 * time.Hour
const requestDataInterval = 10 * time.Minute
const targetDataServerAddr = "10.10.10.12:48080"
const targetRequestPath = "/mempool/api/orangeclock"
const targetDatetimeRequestPath = "/datetime"

func main() {
  for {
    if err := run(); err != nil {
      log.Println("FATAL ERROR:", err)
    }
    log.Println("restart..")
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

  cTimeString, err := httpClient.NewRequest(targetDatetimeRequestPath)
  if err != nil {
    return err
  }
  l := strings.Split(cTimeString, "\n")
  startTime, err := time.Parse(time.RFC3339, l[5])
  if err != nil {
    return err
  }

  // Start
  t := time.Now().Add(displayFullReloadInterval)
  display.UpdateWlanStatus(fmt.Sprintf("#%s", strings.ToUpper(ssid)))
  retryCount := 5
  for {
    logger.Warn("run update cycle")
    if time.Now().After(t) {
      logger.Warn("do a full display reload")
      display.ClearAndSleep()
      t = time.Now().Add(displayFullReloadInterval)
    }

    res, err := httpClient.NewRequest(targetRequestPath)
    if err != nil {
      logger.Error("error while request", slog.String("err", err.Error()))
      retryCount--
    }
    if retryCount <= 0 {
      return errors.New("failed requesting, retries exhausted, restarting")
    }
    logger.Debug("response to display", slog.String("content", res))
    err = drawLines(display, res, startTime)
    if err != nil {
      logger.Error(err.Error())
      retryCount--
    }
    if retryCount <= 0 {
      return errors.New("failed drawing, retries exhausted, restarting")
    }
    time.Sleep(requestDataInterval)
  }

  return nil
}

func drawLines(d *epd2in9v2.PaperDisplay, text string, startTime time.Time) error {
  lines := strings.Split(text, "\n")
  if len(lines) <= 15 {
    return fmt.Errorf("invalid data input, got lines=%d", len(lines))
  }

  d.UpdateLine(lines[5], 0)
  d.UpdateLineMedium(lines[6], 12)
  d.UpdateLine(lines[7], 30)

  d.UpdateLineMedium(lines[9], 47)

  d.UpdateLine(lines[11]+" - "+startTime.Format("02.01.2006 15:04"), 70)
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
  drawLines(d, text, time.Now())

  time.Sleep(60 * time.Second)
}
