package http

import (
	"github.com/soypat/seqs"
	"github.com/soypat/seqs/httpx"
	"github.com/soypat/seqs/stacks"
	"log/slog"
	"math/rand"
	"net/netip"
	"orangeclock/pkg/wifi"
	"time"
)

const connTimeout = 5 * time.Second
const tcpbufsize = 2030 // MTU - ethhdr - iphdr - tcphdr
const hostname = "pico-orangeclock"

type HttpClient struct {
	logger     *slog.Logger
	svAddr     netip.AddrPort
	clientAddr netip.AddrPort
	conn       *stacks.TCPConn
	routerhw   [6]byte
	closeConn  func(err string)
	rng        *rand.Rand
}

func NewHttpClient(logger *slog.Logger, target string) (*HttpClient, string, error) {
	time.Sleep(100 * time.Millisecond)
	dhcpc, stack, _, ssid, err := wifi.SetupWithDHCP(wifi.SetupConfig{
		Hostname: hostname,
		Logger:   logger,
		UDPPorts: 1,
		TCPPorts: 1,
	})
	if err != nil {
		return nil, "", err
	}
	start := time.Now()
	routerhw, err := wifi.ResolveHardwareAddr(stack, dhcpc.Router())
	if err != nil {
		return nil, "", err
	}

	svAddr, err := netip.ParseAddrPort(target)
	if err != nil {
		return nil, "", err
	}

	rng := rand.New(rand.NewSource(int64(time.Now().Sub(start))))
	// Start TCP server.
	clientAddr := netip.AddrPortFrom(stack.Addr(), uint16(rng.Intn(65535-1024)+1024))
	conn, err := stacks.NewTCPConn(stack, stacks.TCPConnConfig{
		TxBufSize: tcpbufsize,
		RxBufSize: tcpbufsize,
	})

	if err != nil {
		return nil, "", err
	}

	closeConn := func(err string) {
		slog.Error("tcpconn:closing", slog.String("err", err))
		conn.Close()
		for !conn.State().IsClosed() {
			slog.Debug("tcpconn:waiting", slog.String("state", conn.State().String()))
			time.Sleep(1000 * time.Millisecond)
		}
	}

	return &HttpClient{
		logger:     logger,
		svAddr:     svAddr,
		clientAddr: clientAddr,
		conn:       conn,
		routerhw:   routerhw,
		closeConn:  closeConn,
		rng:        rng,
	}, ssid, nil
}

func (c *HttpClient) NewRequest(path string) string {
	// Here we create the HTTP request and generate the bytes. The Header method
	// returns the raw header bytes as should be sent over the wire.
	var req httpx.RequestHeader
	req.SetRequestURI(path)
	req.SetMethod("GET")
	req.SetHost(c.svAddr.Addr().String())
	reqbytes := req.Header()

	c.logger.Debug("tcp:ready",
		slog.String("clientaddr", c.clientAddr.String()),
		slog.String("serveraddr", c.svAddr.String()),
	)
	rxBuf := make([]byte, 4096)
	for {
		time.Sleep(5 * time.Second)
		c.logger.Debug("dialing", slog.String("serveraddr", c.svAddr.String()))

		// Make sure to timeout the connection if it takes too long.
		c.conn.SetDeadline(time.Now().Add(connTimeout))
		err := c.conn.OpenDialTCP(c.clientAddr.Port(), c.routerhw, c.svAddr, seqs.Value(c.rng.Intn(65535-1024)+1024))
		if err != nil {
			c.closeConn("opening TCP: " + err.Error())
			continue
		}
		retries := 50
		for c.conn.State() != seqs.StateEstablished && retries > 0 {
			time.Sleep(100 * time.Millisecond)
			retries--
		}
		c.conn.SetDeadline(time.Time{}) // Disable the deadline.
		if retries == 0 {
			c.closeConn("tcp establish retry limit exceeded")
			continue
		}

		// Send the request.
		_, err = c.conn.Write(reqbytes)
		if err != nil {
			c.closeConn("writing request: " + err.Error())
			continue
		}
		time.Sleep(500 * time.Millisecond)
		c.conn.SetDeadline(time.Now().Add(connTimeout))
		n, err := c.conn.Read(rxBuf)
		if n == 0 && err != nil {
			c.closeConn("reading response: " + err.Error())
			continue
		} else if n == 0 {
			c.closeConn("no response")
			continue
		}
		c.logger.Debug("got HTTP response!")
		c.closeConn("done")
		return string(rxBuf[:n])
	}
}
