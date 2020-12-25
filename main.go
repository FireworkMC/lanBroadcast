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

//the default port and address used by vanila minecraft
const broadcastHost = "224.0.2.60:4445"

//LANBroadcast a lan broadcaster
type LANBroadcast struct {
	sender        net.PacketConn
	broadcastAddr *net.UDPAddr
	buffer        bytes.Buffer

	ctx    context.Context
	cancel context.CancelFunc

	mux  sync.Mutex
	port int
	motd string
}

//SetMOTD set the message of the day
func (l *LANBroadcast) SetMOTD(message string) {
	l.mux.Lock()
	l.motd = message
	l.mux.Unlock()
}

//Broadcast broadcasts the game to LAN.
//This function blocks until the given ctx is canceled.
func (l *LANBroadcast) Broadcast() {
	tick := time.Tick(time.Second * 5)
	for {
		select {
		case <-tick:
			if err := l.sendPacket(); err != nil {
				fmt.Println(err)
			}
		case <-l.ctx.Done():
			l.sender.Close()
			return
		}
	}
}

//Close close the LAN broadcaster.
//This function does not block and may be called more than once.
func (l *LANBroadcast) Close() { l.cancel() }

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

//NewLANBroadcast creates a new lan broadcaster
func NewLANBroadcast(ctx context.Context, ip net.IP, port int, motd string) (l *LANBroadcast, err error) {
	if port <= 0 { // disallow negative ports and port 0 since it is not a real port
		return nil, ErrInvalidPort
	}

	if ip == nil || ip.IsUnspecified() || ip.IsLoopback() {
		if ip, err = GetHostAddr(""); err != nil {
			return
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	lan := LANBroadcast{cancel: cancel, ctx: ctx, port: port, motd: motd}
	if lan.sender, err = net.ListenPacket("udp", ip.String()+":"); err == nil {
		if lan.broadcastAddr, err = net.ResolveUDPAddr("udp", broadcastHost); err == nil {
			l = &lan
		}
	}
	return
}
