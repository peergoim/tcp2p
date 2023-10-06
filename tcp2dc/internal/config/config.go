package config

type Config struct {
	StunUrls          []string
	SignalingEndpoint string // http://xxx.xxx.xxx.xx:xxx
	LogLevel          string
	ListenOn          string // 127.0.0.1:10444
	TargetPeerId      string
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
