package server

import (
	"context"
	"encoding/json"
	signaling_client "github.com/peergoim/signaling-client"
	"github.com/peergoim/tcp2p/tcp2dc/internal/config"
	"github.com/pion/webrtc/v2"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type xTcpServer struct {
	Config          *config.Config
	signalingClient *signaling_client.HttpConnection
	ctx             context.Context
	cancel          context.CancelFunc
}

func (s *xTcpServer) startTcpServer() error {
	listenOn := s.Config.ListenOn
	split := strings.Split(listenOn, ":")
	ip := split[0]
	port, _ := strconv.Atoi(split[1])
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
		Zone: "",
	})
	if err != nil {
		signaling_client.Errorf("listen tcp error: %v", err)
		return err
	}
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				conn, err := listener.AcceptTCP()
				if err != nil {
					signaling_client.Errorf("accept tcp error: %v", err)
					continue
				}
				s.handleConn(conn)
			}
		}
	}()
	return nil
}

var TcpServer *xTcpServer
var l sync.Mutex

func SetTcpServer(config *config.Config) error {
	l.Lock()
	defer l.Unlock()
	if TcpServer != nil {
		//重新设置，要把之前的关掉
		TcpServer.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	tmp := &xTcpServer{
		Config: config,
		ctx:    ctx,
		cancel: cancel,
	}
	signalingClient := signaling_client.NewHttpConnection(&signaling_client.Config{
		Endpoint: config.SignalingEndpoint,
		LogLevel: config.LogLevel,
	})
	tmp.signalingClient = signalingClient
	err := tmp.startTcpServer()
	if err != nil {
		return err
	}
	TcpServer = tmp
	go TcpServer.createPeerConnection()
	return nil
}

func (s *xTcpServer) createPeerConnection() *webrtc.PeerConnection {
	// 代理到datachannel
	// 1. 创建PeerConnection
	c := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: s.Config.StunUrls,
			},
		},
	}
	peerConnection, err := webrtc.NewPeerConnection(c)
	if err != nil {
		signaling_client.Errorf("NewPeerConnection error: %v", err)
		return nil
	}
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		signaling_client.Errorf("CreateOffer error: %v", err)
		return nil
	}
	offerBytes, _ := json.Marshal(offer)
	// 发给signaling-server
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	response := s.signalingClient.Call(ctx, &signaling_client.CallRequest{
		PeerId: s.Config.TargetPeerId,
		CallId: "",
		Method: "offer",
		Data:   offerBytes,
	})
	cancel()
	if response.Status != signaling_client.CodeOK {
		signaling_client.Errorf("call signaling-server error: %v", response.Status)
		peerConnection.Close()
		return nil
	}
	// 解析成answer
	answer := webrtc.SessionDescription{}
	err = json.Unmarshal(response.Data, &answer)
	if err != nil {
		signaling_client.Errorf("Unmarshal answer error: %v", err)
		peerConnection.Close()
		return nil
	}
	err = peerConnection.SetRemoteDescription(answer)
	if err != nil {
		signaling_client.Errorf("SetRemoteDescription error: %v", err)
		peerConnection.Close()
		return nil
	}
	signaling_client.Infof("create peer connection success")
	return peerConnection
}

func (s *xTcpServer) handleConn(conn *net.TCPConn) {
	signaling_client.Debugf("handleConn: %v", conn.RemoteAddr())
	defer conn.Close()
	peerConnection := s.createPeerConnection()
	if peerConnection == nil {
		signaling_client.Warnf("handleConn: peerConnection is nil, tcp connection will be closed")
		return
	}
	// 创建datachannel
	stop := make(chan struct{})
	dataChannel, err := peerConnection.CreateDataChannel("data", nil)
	if err != nil {
		signaling_client.Errorf("CreateDataChannel error: %v", err)
		return
	}
	dataChannel.OnOpen(func() {
		signaling_client.Infof("handleConn success")
		for {
			select {
			case <-s.ctx.Done():
				stop <- struct{}{}
				return
			default:
				buf := make([]byte, 1024)
				n, err := conn.Read(buf)
				if err != nil {
					signaling_client.Errorf("tcp read error: %v", err)
					stop <- struct{}{}
					return
				} else if n == 0 {
					signaling_client.Warnf("tcp read 0 bytes")
					continue
				}
				bytes := buf[:n]
				signaling_client.Debugf("tcp read %d bytes", len(bytes))
				dataChannel.Send(bytes)
			}
		}
	})
	// 代理数据
	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		//发给tcp
		_, err := conn.Write(msg.Data)
		if err != nil {
			signaling_client.Errorf("tcp write error: %v", err)
			return
		} else {
			signaling_client.Debugf("tcp write %d bytes", len(msg.Data))
		}
	})
	// onclose
	dataChannel.OnClose(func() {
		stop <- struct{}{}
	})
	<-stop
	dataChannel.Close()
}
