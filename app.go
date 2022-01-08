package main

import (
	"bytes"
	"net"
	"net/http"
	"strings"
)

type App struct {
	AppType      string `json:"appType"`
	ServerHost   string `json:"serverHost"`
	ServerPort   int    `json:"serverPort"`
	ServerSecret string `json:"serverSecret"`
	AdminPort    int    `json:"adminPort"`
	AdminUser    string `json:"adminUser"`
	AdminPass    string `json:"adminPass"`
	TcpPorts     []int  `json:"tcpPorts"`
	UdpPorts     []int  `json:"udpPorts"`

	ip             string
	adminListeners []http.ResponseWriter

	mainListener      net.Listener
	mainConnection    net.Conn
	potentialTcpPorts []int
}

func (a *App) Title() string {
	return "Proxy " + a.AppType
}

func (a *App) ShowServerHost() bool {
	return a.AppType == "client"
}

func (a *App) ShowReconnectButton() bool {
	return a.AppType == "client"
}

func (a *App) ShowRequestPortsButton() bool {
	return a.AppType == "server"
}

func (a *App) PotentialTcpPorts() []int {
	return a.potentialTcpPorts
}

func (a *App) IsConnected() bool {
	return a.mainConnection != nil
}

func (a *App) ConnectionAddr() string {
	if !a.IsConnected() {
		return ""
	}
	return a.mainConnection.RemoteAddr().String()
}

func readMessage(conn net.Conn) (string, bool) {
	if conn == nil {
		return "", false
	}
	buffer := make([]byte, 4096)
	_, err := conn.Read(buffer)
	if err != nil {
		return "", false
	}
	buffer = bytes.Trim(buffer, "\x00")
	if len(buffer) == 0 {
		return "", false
	}
	return strings.TrimSpace(string(buffer)), true
}

func sendMessage(message string, eventFunc func(net.Conn, string)) {
	if app.mainConnection != nil {
		app.mainConnection.Write([]byte(message))
		eventFunc(app.mainConnection, message)
	}
}
