package main

import (
	"encoding/json"
	"net"
	"strconv"
	"strings"
	"sync"
)

var clientMutex sync.Mutex

func openMainConnection() {
	conn, err := net.Dial("tcp", app.ServerHost+":"+strconv.Itoa(app.ServerPort))
	if err != nil {
		logError(err)
		return
	}
	app.mainConnection = conn
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
			processServerMessageConnect(message)
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
		go copyIO(local, proxy)
		go copyIO(proxy, local)
		eventProxyConnection(local, proxy)
	}
}
