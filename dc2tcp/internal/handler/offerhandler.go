package handler

import (
	"context"
	"encoding/json"
	signaling_client "github.com/peergoim/signaling-client"
	"github.com/peergoim/tcp2p/dc2tcp/internal/svc"
	"github.com/pion/sdp/v2"
	webrtc "github.com/pion/webrtc/v2"
	"time"
)

type OfferHandler struct {
	svcCtx *svc.ServiceContext
	req    *signaling_client.CallRequest
}

func NewOfferHandler(
	svcCtx *svc.ServiceContext,
	req *signaling_client.CallRequest,
) *OfferHandler {
	return &OfferHandler{svcCtx: svcCtx, req: req}
}

// Handle webrtc offer handler
func (h *OfferHandler) Handle() *signaling_client.CallResponse {
	var (
		data = h.req.Data
	)
	offer := &webrtc.SessionDescription{}
	err := json.Unmarshal(data, offer)
	if err != nil {
		return &signaling_client.CallResponse{
			CallId: h.req.CallId,
			Method: h.req.Method,
			Status: signaling_client.CodeInvalidArgument,
			Data:   []byte(err.Error()),
		}
	}
	answer, err := h.Logic(offer)
	if err != nil {
		return &signaling_client.CallResponse{
			CallId: h.req.CallId,
			Method: h.req.Method,
			Status: signaling_client.CodeInternal,
			Data:   []byte(err.Error()),
		}
	}
	answerData, err := json.Marshal(answer)
	if err != nil {
		return &signaling_client.CallResponse{
			CallId: h.req.CallId,
			Method: h.req.Method,
			Status: signaling_client.CodeInternal,
			Data:   []byte(err.Error()),
		}
	}
	return &signaling_client.CallResponse{
		CallId: h.req.CallId,
		Method: h.req.Method,
		Status: signaling_client.CodeOK,
		Data:   answerData,
	}
}

func (h *OfferHandler) Logic(in *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: h.svcCtx.Config.StunUrls,
			},
		},
	}
	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		signaling_client.Errorf("webrtc.NewPeerConnection error: %v", err)
		return nil, err
	}
	sd := &sdp.SessionDescription{}
	if err := sd.Unmarshal([]byte(in.SDP)); err != nil {
		signaling_client.Errorf("sd.Unmarshal error: %v", err)
		return nil, err
	}
	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	ctx, cancelFunction := context.WithCancel(context.Background())
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		if connectionState == webrtc.ICEConnectionStateDisconnected {
			cancelFunction()
			signaling_client.Infof("Peer Connection State has changed: %s", connectionState.String())
		} else {
			signaling_client.Infof("Peer Connection State has changed: %s", connectionState.String())
		}
	})

	err = peerConnection.SetRemoteDescription(*in)
	if err != nil {
		signaling_client.Errorf("peerConnection.SetRemoteDescription error: %v", err)
		return nil, err
	}

	// Register data channel creation handling
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		id := d.ID()
		signaling_client.Infof("New DataChannel %s %d", d.Label(), id)

		connection, err := NewConnection(h.svcCtx, ctx, d, sd)
		if err != nil {
			signaling_client.Errorf("NewConnection error: %v", err)
			cancelFunction()
			_ = peerConnection.Close()
			return
		}

		// Register text message handling
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			signaling_client.Infof("Message from DataChannel '%s': '%s'", d.Label(), string(msg.Data))
			for {
				if connection != nil {
					break
				}
				time.Sleep(time.Millisecond * 10)
			}
			connection.OnInput(msg.Data)
		})

		d.OnClose(func() {
			cancelFunction()
			_ = peerConnection.Close()
			signaling_client.Infof("DataChannel '%s'-'%d' closed", d.Label(), id)
		})

	})

	// Create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		signaling_client.Errorf("peerConnection.CreateAnswer error: %v", err)
		return nil, err
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		signaling_client.Errorf("peerConnection.SetLocalDescription error: %v", err)
		return nil, err
	}

	// 返回answer
	return &answer, nil
}
