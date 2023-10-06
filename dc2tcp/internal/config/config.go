package config

type Config struct {
	StunUrls          []string
	TcpAddr           string
	PeerId            string
	SignalingEndpoint string
	LogLevel          string
}

var DefaultStunUrls = []string{
	//"stun:stun.l.google.com:19302",
	//"stun:stun1.l.google.com:19302",
	//"stun:stun2.l.google.com:19302",
	//"stun:stun3.l.google.com:19302",
	//"stun:stun4.l.google.com:19302",
	// 中国节点
	"stun:stun.newrocktech.com",
	"stun:stun.qq.com",
	"stun:stun.miwifi.com",
}
