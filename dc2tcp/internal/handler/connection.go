package handler

import (
	"context"
	signaling_client "github.com/peergoim/signaling-client"
	"github.com/peergoim/tcp2p/dc2tcp/internal/svc"
	"github.com/pion/sdp/v2"
	"github.com/pion/webrtc/v2"
	proxyproto "github.com/pires/go-proxyproto"
	"net"
	"strconv"
	"strings"
)

type Connection struct {
	svcCtx      *svc.ServiceContext
	DataChannel *webrtc.DataChannel
	TcpConn     *net.TCPConn
	SDP         *sdp.SessionDescription
	ctx         context.Context
}

func getClientAddr(sd *sdp.SessionDescription) (string, int) {
	for _, m := range sd.MediaDescriptions {
		address := m.ConnectionInformation.Address.Address
		split := strings.Split(address, ":")
		if len(split) == 2 {
			port, _ := strconv.Atoi(split[1])
			return split[0], port
		}
		return split[0], 0
	}
	return "", 0
}

func NewConnection(svcCtx *svc.ServiceContext, ctx context.Context, dc *webrtc.DataChannel, sd *sdp.SessionDescription) (*Connection, error) {
	// 建立新的tcp连接
	tcpAddr := svcCtx.Config.TcpAddr
	split := strings.Split(tcpAddr, ":")
	tcpIp := split[0]
	tcpPort, _ := strconv.Atoi(split[1])

	tcpConn, err := net.Dial("tcp", tcpAddr)
	if err != nil {
		signaling_client.Errorf("dial tcp error: %v", err)
		return nil, err
	}
	clientIp, clientPort := getClientAddr(sd)
	header := &proxyproto.Header{
		Version:           1,
		Command:           proxyproto.PROXY,
		TransportProtocol: proxyproto.TCPv4,
		SourceAddr: &net.TCPAddr{
			IP:   net.ParseIP(clientIp),
			Port: clientPort,
		},
		DestinationAddr: &net.TCPAddr{
			IP:   net.ParseIP(tcpIp),
			Port: tcpPort,
		},
	}
	// After the connection was created write the proxy headers first
	_, _ = header.WriteTo(tcpConn)
	//把tcp读到的数据原封不动的写到datachannel里面
	c := &Connection{
		svcCtx:      svcCtx,
		DataChannel: dc,
		TcpConn:     tcpConn.(*net.TCPConn),
		SDP:         sd,
		ctx:         ctx,
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				c.OnClose()
				return
			default:
				buf := make([]byte, 1024)
				n, err := tcpConn.Read(buf)
				if err != nil {
					signaling_client.Errorf("tcp read error: %v", err)
					c.OnClose()
					return
				}
				dc.Send(buf[:n])
			}
		}
	}()
	return c, nil
}

func (c *Connection) OnInput(data []byte) {
	c.TcpConn.Write(data)
}

func (c *Connection) OnClose() {
	_ = c.TcpConn.Close()
	_ = c.DataChannel.Close()
}
