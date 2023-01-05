package lanbroadcast

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/yehan2002/errors"
)

const (
	//ErrInvalidPort invalid port provided
	ErrInvalidPort = errors.Error("lanBroadcast: Invalid port provided")
	//ErrHost unable to get the host address
	ErrHost = errors.Error("lanBroadcast: unable to get the host address or interface")
)

// the default port and address used by vanilla minecraft
const broadcastHost = "224.0.2.60:4445"

// LANBroadcast a lan broadcaster
type LANBroadcast struct {
	sender        net.PacketConn
	broadcastAddr *net.UDPAddr
	buffer        bytes.Buffer
	interval      time.Duration

	cancelMux sync.Mutex
	cancel    context.CancelFunc

	mux  sync.Mutex
	port int
	motd string
}

// SetMOTD set the message of the day
func (l *LANBroadcast) SetMOTD(message string) {
	l.mux.Lock()
	l.motd = message
	l.mux.Unlock()
}

// SetInterval set the interval between broadcasts in seconds.
// This has no effect if called after calling `Broadcast()`.
func (l *LANBroadcast) SetInterval(t uint) {
	l.mux.Lock()
	l.interval = time.Second * time.Duration(t)
	l.mux.Unlock()
}

// Broadcast broadcasts the game to LAN.
// This function blocks until the given ctx is canceled or Close is called.
func (l *LANBroadcast) Broadcast(ctx context.Context) {
	var interval = time.Second * 5
	if l.interval != 0 {
		interval = l.interval
	}

	l.cancelMux.Lock()
	if l.cancel != nil {
		panic("Tried to start multiple broadcasts")
	}
	ctx, l.cancel = context.WithCancel(ctx)
	l.cancelMux.Unlock()

	tick := time.Tick(interval)
	for {
		select {
		case <-tick:
			if err := l.sendPacket(); err != nil {
				fmt.Println("Error sending LAN broadcast:", err)
			}
		case <-ctx.Done():
			l.sender.Close()
			return
		}
	}
}

// Close close the LAN broadcaster.
// This function does not block and may be called more than once.
func (l *LANBroadcast) Close() {
	l.cancelMux.Lock()
	if l.cancel != nil {
		l.cancel()
		l.cancel = nil
	}
	l.cancelMux.Unlock()
}

func (l *LANBroadcast) sendPacket() error {
	l.mux.Lock()
	l.buffer.Reset()
	fmt.Fprintf(&l.buffer, "[MOTD]%s[/MOTD][AD]%d[/AD]", l.motd, l.port)
	l.mux.Unlock()
	n, err := l.sender.WriteTo(l.buffer.Bytes(), l.broadcastAddr)
	if n != l.buffer.Len() && err == nil {
		return io.ErrShortWrite
	}
	return err
}

// New creates a new lan broadcaster
func New(ip net.IP, port int, motd string) (l *LANBroadcast, err error) {
	if port <= 0 { // disallow negative ports and port 0 since it is not a real port
		return nil, ErrInvalidPort
	}

	if ip == nil || ip.IsUnspecified() || ip.IsLoopback() {
		if ip, err = getHostAddr(""); err != nil {
			return
		}
	}

	lan := LANBroadcast{port: port, motd: motd}
	if lan.sender, err = net.ListenPacket("udp", ip.String()+":"); err == nil {
		if lan.broadcastAddr, err = net.ResolveUDPAddr("udp", broadcastHost); err == nil {
			l = &lan
		}
	}
	return
}
