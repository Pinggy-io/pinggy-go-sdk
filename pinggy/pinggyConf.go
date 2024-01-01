package pinggy

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh"
)

func (conf *Config) verify() {
	if conf.Server == "" {
		conf.Server = "a.pinggy.io"
	}
	addr := strings.Split(conf.Server, ":")
	conf.port = 443
	conf.Server = addr[0]
	if len(addr) > 1 {
		p, err := strconv.Atoi(addr[1])
		if err != nil {
			conf.logger.Fatal(err)
		}
		conf.port = p
	}
	if conf.logger == nil {
		conf.logger = log.Default()
	}

	ctype := conf.Type
	switch ctype {
	case HTTP, TCP, TLS:
		conf.Type = ctype
	default:
		conf.Type = ""
	}
	atype := conf.AltType
	conf.AltType = ""
	switch atype {
	case UDP:
		conf.AltType = UDP
	default:
		conf.AltType = ""
	}

	if conf.Type != "" && conf.AltType != "" {
		conf.AltType = ""
	}

	if conf.Type == "" && conf.AltType == "" {
		conf.Type = HTTP
	}
}

func dialWithConfig(conf *Config) (*ssh.Client, error) {
	user := "auth"
	if conf.Type != "" {
		user += "+" + string(conf.Type)
	}
	if conf.AltType != "" {
		user += "+" + string(conf.AltType)
	}
	if conf.Token != "" {
		user = conf.Token + "+" + user
	}
	clientConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password("nopass"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	usingToken := "without using any token"
	if conf.Token != "" {
		usingToken = fmt.Sprintf("using token: %s", conf.Token)
	}
	conf.logger.Printf("Initiating ssh connection %s to server: %s:%d\n", usingToken, conf.Server, conf.port)

	addr := fmt.Sprintf("%s:%d", conf.Server, conf.port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		conf.logger.Printf("Error in ssh connection initiation: %v\n", err)
		return nil, err
	}
	if conf.SshOverSsl {
		tlsConn := tls.Client(conn, &tls.Config{ServerName: conf.Server})
		err := tlsConn.Handshake()
		if err != nil {
			conf.logger.Printf("Error in ssh connection initiation: %v\n", err)
			return nil, err
		}
		conn = tlsConn
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, clientConfig)
	if err != nil {
		return nil, err
	}

	return ssh.NewClient(c, chans, reqs), nil
}
