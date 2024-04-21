package epd2in9v2

import (
	"log/slog"
	"machine"
	"strings"
	"time"
)

const (
	width  = 128
	height = 296
)

type PaperDisplay struct {
	Display Device
	Status  string
	logger  *slog.Logger
}

func NewPaperDisplay(logger *slog.Logger) *PaperDisplay {
	err := machine.SPI1.Configure(machine.SPIConfig{
		Frequency: 4_000_000,
		Mode:      0,
	})
	if err != nil {
		logger.Error(err.Error())
	}

	csPin := machine.GP9
	dcPin := machine.GP8
	rstPin := machine.GP12
	busyPin := machine.GP13

	display := New(machine.SPI1, csPin, dcPin, rstPin, busyPin)
	display.Configure(Config{
		Width:        width,
		Height:       height,
		LogicalWidth: width,
		Rotation:     ROTATION_180,
	})

	display.Init()
	display.Clear()
	time.Sleep(2 * time.Second)

	display.Sleep()
	time.Sleep(2 * time.Second)
	return &PaperDisplay{
		Display: display,
		logger:  logger,
	}
}

func (d *PaperDisplay) UpdateRawText(text string) {
	lines := strings.Split(text, "\n")
	d.logger.Debug("got lines", slog.Int("count", len(lines)))
	for i, l := range lines {
		if l == "" || l == "\n" {
			continue
		}
		d.Display.DrawStringSmall(int16(i*10), height-10, l)
		d.logger.Debug("draw line at", slog.Int("pos", i*10))
	}
	d.Display.DisplayPartial()
}

func (d *PaperDisplay) UpdateLine(line string, x int) {
	d.Display.DrawStringSmall(int16(x), height-10, line)
	d.logger.Debug("draw line at", slog.Int("pos", x))
	d.Display.DisplayPartial()
}

func (d *PaperDisplay) UpdateLineMedium(line string, x int) {
	d.Display.DrawStringMedium(int16(x), height-10, line)
	d.logger.Debug("draw medium line at", slog.Int("pos", x))
	d.Display.DisplayPartial()
}

func (d *PaperDisplay) UpdateWlanStatus(status string) {
	if d.Status != status {
		d.Display.DrawStringSmall(0, 60, status)
		d.logger.Debug("update status on display")
	}
	d.Display.DisplayPartial()
	d.Status = status
}

func (d *PaperDisplay) ClearAndSleep() {
	d.Display.Clear()
	time.Sleep(2 * time.Second)
	d.Display.Sleep()
	time.Sleep(2 * time.Second)
}
