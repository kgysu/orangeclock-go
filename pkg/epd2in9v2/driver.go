package epd2in9v2

import (
	"fmt"
	"image/color"
	"machine"
	font_medium "orangeclock/pkg/font-medium"
	font_small "orangeclock/pkg/font-small"
	"time"
	"tinygo.org/x/drivers"
)

var (
	Black = color.RGBA{1, 1, 1, 255}
	White = color.RGBA{0, 0, 0, 255}
)

type Config struct {
	Width        int16 // Width is the display resolution
	Height       int16
	LogicalWidth int16    // LogicalWidth must be a multiple of 8 and same size or bigger than Width
	Rotation     Rotation // Rotation is clock-wise
}

type Device struct {
	bus          drivers.SPI
	cs           machine.Pin
	dc           machine.Pin
	rst          machine.Pin
	busy         machine.Pin
	logicalWidth int16
	width        int16
	height       int16
	buffer       []uint8
	bufferLength uint32
	rotation     Rotation
}

type Rotation uint8

var lutWF_Partial = [159]uint8{
	0x0, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x80, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x40, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0A, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2,
	0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x0, 0x0, 0x0,
	0x22, 0x17, 0x41, 0xB0, 0x32, 0x36,
}

var lutWS_20_30 = [159]uint8{
	0x80, 0x66, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x0, 0x0, 0x0,
	0x10, 0x66, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x20, 0x0, 0x0, 0x0,
	0x80, 0x66, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x0, 0x0, 0x0,
	0x10, 0x66, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x20, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x14, 0x8, 0x0, 0x0, 0x0, 0x0, 0x1,
	0xA, 0xA, 0x0, 0xA, 0xA, 0x0, 0x1,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x14, 0x8, 0x0, 0x1, 0x0, 0x0, 0x1,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x0, 0x0, 0x0,
	0x22, 0x17, 0x41, 0x0, 0x32, 0x36,
}

var lutGray4 = [159]uint8{
	0x00, 0x60, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //VS L0	 //2.28s
	0x20, 0x60, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //VS L1
	0x28, 0x60, 0x14, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //VS L2
	0x2A, 0x60, 0x15, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //VS L3
	0x00, 0x90, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //VS L4
	0x00, 0x02, 0x00, 0x05, 0x14, 0x00, 0x00, //TP, SR, RP of Group0
	0x1E, 0x1E, 0x00, 0x00, 0x00, 0x00, 0x01, //TP, SR, RP of Group1
	0x00, 0x02, 0x00, 0x05, 0x14, 0x00, 0x00, //TP, SR, RP of Group2
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //TP, SR, RP of Group3
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //TP, SR, RP of Group4
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //TP, SR, RP of Group5
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //TP, SR, RP of Group6
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //TP, SR, RP of Group7
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //TP, SR, RP of Group8
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //TP, SR, RP of Group9
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //TP, SR, RP of Group10
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //TP, SR, RP of Group11
	0x24, 0x22, 0x22, 0x22, 0x23, 0x32, 0x00, 0x00, 0x00, //FR, XON
	0x22, 0x17, 0x41, 0xAE, 0x32, 0x28, //EOPT VGH VSH1 VSH2 VSL VCOM
}

// New returns a new epd2in9 driver. Pass in a fully configured SPI bus.
func New(bus drivers.SPI, csPin, dcPin, rstPin, busyPin machine.Pin) Device {
	csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	dcPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	rstPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	busyPin.Configure(machine.PinConfig{Mode: machine.PinInput})
	return Device{
		bus:  bus,
		cs:   csPin,
		dc:   dcPin,
		rst:  rstPin,
		busy: busyPin,
	}
}

func (d *Device) Configure(cfg Config) {
	if cfg.LogicalWidth != 0 {
		d.logicalWidth = cfg.LogicalWidth
	} else {
		d.logicalWidth = 128
	}
	if cfg.Width != 0 {
		d.width = cfg.Width
	} else {
		d.width = 128
	}
	if cfg.Height != 0 {
		d.height = cfg.Height
	} else {
		d.height = 296
	}
	d.rotation = cfg.Rotation
	d.bufferLength = (uint32(d.logicalWidth) * uint32(d.height)) / 8
	d.buffer = make([]uint8, d.bufferLength)
	for i := uint32(0); i < d.bufferLength; i++ {
		d.buffer[i] = 0xFF
	}
}

// Reset Software reset
func (d *Device) Reset() {
	d.rst.High()
	time.Sleep(10 * time.Millisecond)
	d.rst.Low()
	time.Sleep(2 * time.Millisecond)
	d.rst.High()
	time.Sleep(10 * time.Millisecond)
}

// SendCommand sends a command to the display
func (d *Device) SendCommand(command uint8) {
	d.dc.Low()
	d.cs.Low()
	_, err := d.bus.Transfer(command)
	if err != nil {
		println(err)
	}
	d.cs.High()
}

// SendData sends a data byte to the display
func (d *Device) SendData(data uint8) {
	d.dc.High()
	d.cs.Low()
	_, err := d.bus.Transfer(data)
	if err != nil {
		println(err)
	}
	d.cs.High()
}

// ReadBusy waits until the busy_pin goes LOW
func (d *Device) ReadBusy() {
	for d.busy.Get() {
		time.Sleep(50 * time.Millisecond)
	}
	time.Sleep(50 * time.Millisecond)
}

func (d *Device) Lut(lut [159]uint8) {
	d.SendCommand(WRITE_LUT_REGISTER)
	for i := 0; i < 153; i++ {
		d.SendData(lut[i])
	}
	d.ReadBusy()
}

func (d *Device) LutByHost(lut [159]uint8) {
	d.Lut(lut)
	d.SendCommand(0x3f)
	d.SendData(lut[153])
	d.SendCommand(SET_GATE_DRIVING_VOLTAGE) // gate voltage
	d.SendData(lut[154])
	d.SendCommand(0x04)  // source voltage
	d.SendData(lut[155]) // VSH
	d.SendData(lut[156]) // VSH2
	d.SendData(lut[157]) // VSL
	d.SendCommand(0x2c)  // VCOM
	d.SendData(lut[158])
}

func (d *Device) TurnOnDisplay() {
	d.SendCommand(DISPLAY_UPDATE_CONTROL_2)
	d.SendData(0xc7)
	d.SendCommand(MASTER_ACTIVATION)
	d.ReadBusy()
}

func (d *Device) TurnOnDisplayPartial() {
	d.SendCommand(DISPLAY_UPDATE_CONTROL_2)
	d.SendData(0x0F)
	d.SendCommand(MASTER_ACTIVATION)
	d.ReadBusy()
}

// setWindows setting the display window
func (d *Device) setWindows(x0 int16, y0 int16, x1 int16, y1 int16) {
	d.SendCommand(SET_RAM_X_ADDRESS_START_END_POSITION)
	d.SendData(uint8((x0 >> 3) & 0xFF))
	d.SendData(uint8((x1 >> 3) & 0xFF))
	d.SendCommand(SET_RAM_Y_ADDRESS_START_END_POSITION)
	d.SendData(uint8(y0 & 0xFF))
	d.SendData(uint8((y0 >> 8) & 0xFF))
	d.SendData(uint8(y1 & 0xFF))
	d.SendData(uint8((y1 >> 8) & 0xFF))
}

// setCursor moves the internal pointer to the specified coordinates
func (d *Device) setCursor(x int16, y int16) {
	d.SendCommand(SET_RAM_X_ADDRESS_COUNTER)
	d.SendData(uint8(x & 0xFF)) // TODO check from SetCursor
	//d.SendData(uint8((x >> 3) & 0xFF))

	d.SendCommand(SET_RAM_Y_ADDRESS_COUNTER)
	d.SendData(uint8(y & 0xFF))
	d.SendData(uint8((y >> 8) & 0xFF))
	d.ReadBusy()
}

// Init initialize the e-paper register
func (d *Device) Init() {
	d.Reset()
	time.Sleep(100 * time.Millisecond)

	d.ReadBusy()
	d.SendCommand(0x12)
	d.ReadBusy()

	d.SendCommand(DRIVER_OUTPUT_CONTROL)
	d.SendData(0x27)
	d.SendData(0x01)
	d.SendData(0x00)

	d.SendCommand(DATA_ENTRY_MODE_SETTING)
	d.SendData(0x03)

	d.setWindows(0, 0, d.width-1, d.height-1)

	d.SendCommand(DISPLAY_UPDATE_CONTROL_1)
	d.SendData(0x00)
	d.SendData(0x80)

	d.setCursor(0, 0)
	d.ReadBusy()

	d.LutByHost(lutWS_20_30)
}

func (d *Device) Gray4Init() {
	d.Reset()
	time.Sleep(100 * time.Millisecond)

	d.ReadBusy()
	d.SendCommand(0x12) // soft reset
	d.ReadBusy()

	d.SendCommand(DRIVER_OUTPUT_CONTROL)
	d.SendData(0x27)
	d.SendData(0x01)
	d.SendData(0x00)

	d.SendCommand(DATA_ENTRY_MODE_SETTING)
	d.SendData(0x03)

	d.setWindows(8, 0, d.width, d.height-1)

	d.SendCommand(0x3C)
	d.SendData(0x04)

	d.setCursor(1, 0)
	d.ReadBusy()

	d.LutByHost(lutGray4)
}

// Clear clears the screen
func (d *Device) Clear() {
	d.SendCommand(WRITE_RAM)
	for i := 0; i < 4736; i++ {
		d.SendData(0xFF)
	}

	d.SendCommand(0x26)
	for i := 0; i < 4736; i++ {
		d.SendData(0xFF)
	}
	d.TurnOnDisplay()
}

// Display Sends the image buffer in RAM to e-Paper and displays
func (d *Device) Display() {
	d.SendCommand(WRITE_RAM)
	for i := 0; i < 4736; i++ {
		d.SendData(d.buffer[i])
	}
	d.TurnOnDisplay()
}

func (d *Device) DisplayBase() {
	d.SendCommand(WRITE_RAM)
	for i := 0; i < 4736; i++ {
		d.SendData(d.buffer[i])
	}
	d.SendCommand(0x26)
	for i := 0; i < 4736; i++ {
		d.SendData(d.buffer[i])
	}
	d.TurnOnDisplay()
}

func (d *Device) Gray4Display() {
	// todo implement
}

func (d *Device) DisplayPartial() {
	// reset
	d.rst.Low()
	time.Sleep(time.Millisecond)
	d.rst.High()
	time.Sleep(2 * time.Millisecond)

	d.Lut(lutWF_Partial)
	d.SendCommand(0x37)
	d.SendData(0x00)
	d.SendData(0x00)
	d.SendData(0x00)
	d.SendData(0x00)
	d.SendData(0x00)
	d.SendData(0x40)
	d.SendData(0x00)
	d.SendData(0x00)
	d.SendData(0x00)
	d.SendData(0x00)

	d.SendCommand(0x3C) // BorderWavefrom
	d.SendData(0x80)

	d.SendCommand(0x22)
	d.SendData(0xC0)
	d.SendCommand(0x20)
	d.ReadBusy()

	d.setWindows(0, 0, d.width-1, d.height-1)
	d.setCursor(0, 0)

	d.SendCommand(0x24) // write Black and White image to RAM
	for i := 0; i < 4736; i++ {
		d.SendData(d.buffer[i])
	}
	d.TurnOnDisplayPartial()
}

func (d *Device) Sleep() {
	d.SendCommand(DEEP_SLEEP_MODE)
	d.SendData(0x01)
	time.Sleep(100 * time.Millisecond)
}

// Paint

func (d *Device) DrawStringSmall(startX, startY int16, text string) {
	for rN, r := range text {
		runeData := font_small.RuneToBitmapFontSmall(r)
		nextCharPos := int16(rN*6) + int16(rN*1) // +1 for spacing
		for x, rd := range runeData {
			for y, p := range rd {
				if p == 1 {
					d.SetPixel(startX+int16(x), startY-nextCharPos-int16(y), Black)
				} else {
					d.SetPixel(startX+int16(x), startY-nextCharPos-int16(y), White)
				}
			}
		}
	}
}

func (d *Device) DrawStringMedium(startX, startY int16, text string) {
	for rN, r := range text {
		runeData := font_medium.RuneToBitmapMedium(r)
		nextCharPos := int16(rN*12) + int16(rN*1) // +1 for spacing
		for x, rd := range runeData {
			for y, p := range rd {
				if p == 1 {
					d.SetPixel(startX+int16(x), startY-nextCharPos-int16(y), Black)
				} else {
					d.SetPixel(startX+int16(x), startY-nextCharPos-int16(y), White)
				}
			}
		}
	}
}

func (d *Device) DrawRectangle(x, y, width, height int) {
	// bottom Line
	d.DrawHorizontalLine(x, y, width)
	// top Line
	d.DrawHorizontalLine(x+height-1, y, width)
	d.DrawVerticalLine(x, y, height)
	d.DrawVerticalLine(x, y+width-1, height)
}

func (d *Device) DrawVerticalLine(x, y, length int) {
	for i := 0; i < length; i++ {
		d.SetPixelDefault(x+i, y)
	}
}

func (d *Device) DrawHorizontalLine(x, y, length int) {
	for i := 0; i < length; i++ {
		d.SetPixelDefault(x, y+i)
	}
}

// EXTRAS

func (d *Device) SetPixelDefault(x int, y int) {
	d.SetPixel(int16(x), int16(y), Black)
}

// SetPixel modifies the internal buffer in a single pixel.
// The display have 2 colors: black and white
// We use RGBA(0,0,0, 255) as white (transparent)
// Anything else as black
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	x, y = d.xy(x, y)
	if x < 0 || x >= d.logicalWidth || y < 0 || y >= d.height {
		fmt.Printf("Drawing out of space, width=%d, x=%d; height=%d, y=%d\n", d.width, x, d.height, y)
		return
	}
	byteIndex := (int32(x) + int32(y)*int32(d.logicalWidth)) / 8
	if c.R == 0 && c.G == 0 && c.B == 0 { // TRANSPARENT / WHITE
		d.buffer[byteIndex] |= 0x80 >> uint8(x%8)
	} else { // WHITE / EMPTY
		d.buffer[byteIndex] &^= 0x80 >> uint8(x%8)
	}
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	if d.rotation == ROTATION_90 || d.rotation == ROTATION_270 {
		return d.height, d.logicalWidth
	}
	return d.logicalWidth, d.height
}

// SetRotation changes the rotation (clock-wise) of the device
func (d *Device) SetRotation(rotation Rotation) {
	d.rotation = rotation
}

// xy chages the coordinates according to the rotation
func (d *Device) xy(x, y int16) (int16, int16) {
	switch d.rotation {
	case NO_ROTATION:
		return x, y
	case ROTATION_90:
		return d.width - y - 1, x
	case ROTATION_180:
		return d.width - x - 1, d.height - y - 1
	case ROTATION_270:
		return y, d.height - x - 1
	}
	return x, y
}
