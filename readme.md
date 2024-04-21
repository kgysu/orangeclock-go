# OrangeClock Go

Own implementation of the OrangeClock written in Go (tinygo).

It uses a Raspberry Pico with Wi-Fi and a Waveshare E-Paper display.
Hardware credits go to: [OrangeClock](https://orange-clock.com/)

Also it needs a webserver in the local network which serves the content to display.
Something like this: [server](https://github.com/kgysu/orangeclock-server)



## Flashing

Before you flash make sure to enter your wifi-credentials in `pkg/wifi/secrets.go`.

To flash the app to the raspi-pico_w run:

```bash
tinygo flash -target=pico -stack-size=8kb -monitor ./cmd/orangeclock/main.go

# Monitor
tinygo monitor
```



## Resources

- [OrangeClock hardware](https://orange-clock.com/)
- [Raspberry pico-w](https://www.raspberrypi.com/documentation/microcontrollers/raspberry-pi-pico.html#raspberry-pi-pico-w)
- [tinygo](https://tinygo.org/)
- [Wifi chip CYW43439 docs](https://www.infineon.com/dgdl/Infineon-CYW43439-DataSheet-v05_00-EN.pdf)
- [Wifi Chip driver](https://github.com/soypat/cyw43439)
- [e-paper display](https://www.waveshare.com/pico-epaper-2.9.htm)
- [epd2in9 docs](https://www.waveshare.com/wiki/Pico-ePaper-2.9)
- [Mempool data](https://mempool.space/)


