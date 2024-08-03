package main

import (
	"encoding/json"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var clientMutex sync.Mutex

func openMainConnection() {
	for {
		conn, err := net.Dial("tcp", app.ServerHost+":"+strconv.Itoa(app.ServerPort))
		if err != nil {
			logError(err)
			eventLog("Server not found, will try to reconnect in 10 seconds")
		} else {
			app.mainConnection = conn
			break
		}
		time.Sleep(10 * time.Second)
	}
	eventServerConnectionStarted(app.mainConnection)
	sendServerSecret()
	sendTcpPorts()
	readFromMainConnection()
}

func closeMainConnection() {
	clientMutex.Lock()
	defer clientMutex.Unlock()
	if app.mainConnection != nil {
		eventServerConnectionEnded(app.mainConnection)
		app.mainConnection.Close()
		app.mainConnection = nil
	}
}

func sendMessageToServer(message string) {
	sendMessage(message, eventMessageSentToServer)
}

func sendServerSecret() {
	sendMessageToServer("secret=" + app.ServerSecret)
}

func sendTcpPorts() {
	s, _ := json.Marshal(app.TcpPorts)
	sendMessageToServer("tcpports=" + strings.Trim(string(s), "[]"))
}

func reloadOpenTcpPorts() {
	app.potentialTcpPorts = []int{}
	for i := 1; i < 65536; i++ {
		if isTcpPortOpen(i) {
			app.potentialTcpPorts = append(app.potentialTcpPorts, i)
		}
	}
}

func readFromMainConnection() {
	for {
		if message, ok := readMessage(app.mainConnection); ok {
			eventMessageReceivedFromServer(app.mainConnection, message)
			processServerMessageTcpPorts(message)
			go processServerMessageConnect(message)
		} else {
			closeMainConnection()
			break
		}
	}
}

func processServerMessageTcpPorts(message string) {
	if message == "tcpports" {
		sendTcpPorts()
	}
}

func processServerMessageConnect(message string) {
	if network, localPort, proxyPort, ok := getConnectionFromMessage(message); ok {
		local, err := net.Dial(network, "127.0.0.1:"+strconv.Itoa(localPort))
		if err != nil {
			logError(err)
		}
		proxy, err := net.Dial(network, app.ServerHost+":"+strconv.Itoa(proxyPort))
		if err != nil {
			logError(err)
		}
		eventProxyConnection(local, proxy)
		if network == "udp" {
			go proxy.Write([]byte("connect"))
			for {
				copyIO(local, proxy)
			}
		} else {
			copyIO(local, proxy)
		}
	}
}
