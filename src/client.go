package main

import (
	"net"
	"strconv"
	"sync"
	"time"
)

var clientMutex sync.Mutex

func openMainConnection() {
	for {
		conn, err := net.Dial("tcp", app.ServerHost+":"+strconv.Itoa(app.ServerPort))
		if err != nil {
			logError(err)
		} else {
			app.mainConnection = conn
			logInfo("Connected to the server successfully...")
			break
		}
		logInfo("Will try to reconnect in 10 seconds...")
		time.Sleep(10 * time.Second)
	}
	eventServerConnectionStarted(app.mainConnection)
	time.Sleep(100 * time.Millisecond)
	sendServerSecret()
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

func readFromMainConnection() {
	for {
		if message, ok := readMessage(app.mainConnection); ok {
			eventMessageReceivedFromServer(app.mainConnection, message)
			go processServerMessageConnect(message)
		} else {
			closeMainConnection()
			break
		}
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
			proxy.Write([]byte("connect"))
			for {
				copyIO(local, proxy)
			}
		} else {
			copyIO(local, proxy)
		}
	}
}
