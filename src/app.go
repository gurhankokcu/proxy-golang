package main

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type App struct {
	AppType      string   `json:"appType"`
	ServerHost   string   `json:"serverHost"`
	ServerPort   int      `json:"serverPort"`
	ServerSecret string   `json:"serverSecret"`
	AdminPort    int      `json:"adminPort"`
	AdminUser    string   `json:"adminUser"`
	AdminPass    string   `json:"adminPass"`
	TcpPorts     []int    `json:"tcpPorts"`
	UdpPorts     []int    `json:"udpPorts"`
	Events       []string `json:"events"`

	Ip                   string
	potentialTcpPorts    []int
	adminListeners       []http.ResponseWriter
	mainListener         net.Listener
	mainConnection       net.Conn
	userTcpListeners     map[string]*UserTcpListener
	clientUdpConnections map[string]*ClientUdpConnection
}

type UserTcpListener struct {
	listener    net.Listener
	connections []*UserTcpConnection
}

type UserTcpConnection struct {
	connection       net.Conn
	clientListener   net.Listener
	clientConnection net.Conn
}

type ClientUdpConnection struct {
	connection      *net.UDPConn
	remoteAddr      *net.UDPAddr
	userConnections []*UserUdpConnection
}

type UserUdpConnection struct {
	connection *net.UDPConn
	remoteAddr *net.UDPAddr
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
	ports := []int{}
	for _, port := range a.potentialTcpPorts {
		if indexOfItemInIntSlice(&a.TcpPorts, port) == -1 {
			ports = append(ports, port)
		}
	}
	a.potentialTcpPorts = ports
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
		time.Sleep(100 * time.Millisecond)
	}
}

func copyIO(src, dest net.Conn) {
	defer dest.Close()
	defer src.Close()

	go func() {
		_, err := io.Copy(dest, src)
		if err != nil {
			logError(err)
		}
	}()

	_, err := io.Copy(src, dest)
	if err != nil {
		logError(err)
	}
}

func udpCopyIO(userUdpConnection *UserUdpConnection, clientUdpConnection *ClientUdpConnection) {
	buffer := make([]byte, 4096)
	for {
		n, clientAddr, err := userUdpConnection.connection.ReadFromUDP(buffer)
		if err != nil {
			logError(err)
			return
		}
		logInfo("read from user")
		userUdpConnection.remoteAddr = clientAddr

		go func(userUdpConnection *UserUdpConnection, clientUdpConnection *ClientUdpConnection) {
			response := make([]byte, 4096)
			for {
				n, _, err = clientUdpConnection.connection.ReadFromUDP(response)
				if err != nil {
					logError(err)
					return
				}
				logInfo("read from client")
				_, err = userUdpConnection.connection.WriteToUDP(response[:n], userUdpConnection.remoteAddr)
				if err != nil {
					logError(err)
					return
				}
				logInfo("sent to user")
			}
		}(userUdpConnection, clientUdpConnection)

		_, err = clientUdpConnection.connection.WriteToUDP(buffer[:n], clientUdpConnection.remoteAddr)
		if err != nil {
			logError(err)
			return
		}
		logInfo("sent to client")

	}
}
