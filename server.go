package main

import (
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var serverMutex sync.Mutex

func openMainListener() {
	l, err := net.Listen("tcp", ":"+strconv.Itoa(app.ServerPort))
	if err != nil {
		logError(err)
		return
	}
	app.mainListener = l
	listenMainListenerConnections()
}

func closeMainListener() {
	serverMutex.Lock()
	defer serverMutex.Unlock()
	if app.mainConnection != nil {
		app.mainConnection.Close()
		app.mainConnection = nil
	}
	if app.mainListener != nil {
		app.mainListener.Close()
		app.mainListener = nil
	}
}

func sendMessageToClient(message string) {
	sendMessage(message, eventMessageSentToClient)
}

func requestOpenTcpPorts() {
	sendMessageToClient("tcpports")
}

func listenMainListenerConnections() {
	for {
		if app.mainListener == nil {
			return
		}
		conn, err := app.mainListener.Accept()
		if err != nil {
			logError(err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		eventClientConnectionStarted(conn)
		if app.mainConnection != nil {
			closeMainListenerConnection(conn, eventClientConnectionRejectedAlreadyConnected)
			continue
		}
		checkMainListenerConnection(conn)
		go handleMainListenerConnection(conn)
	}
}

func checkMainListenerConnection(conn net.Conn) {
	time.AfterFunc(10*time.Second, func() {
		if app.mainConnection != conn {
			closeMainListenerConnection(conn, eventClientConnectionRejectedTimeout)
		}
	})
}

func handleMainListenerConnection(conn net.Conn) {
	for {
		if message, ok := readMessage(conn); ok {
			eventMessageReceivedFromClient(conn, message)
			processClientMessageServerSecret(conn, message)
			processClientMessageTcpPorts(conn, message)
		} else {
			isMainConnection := conn == app.mainConnection
			closeMainListenerConnection(conn, eventClientConnectionEnded)
			if isMainConnection {
				app.mainConnection = nil
			}
			break
		}
	}
}

func processClientMessageServerSecret(conn net.Conn, message string) {
	if conn == app.mainConnection {
		return
	}
	if serverSecret := getServerSecretFromMessage(message); serverSecret != "" {
		if serverSecret != app.ServerSecret {
			closeMainListenerConnection(conn, eventClientConnectionRejectedInvalidSecret)
			return
		}
		if app.mainConnection != nil {
			closeMainListenerConnection(conn, eventClientConnectionRejectedAlreadyConnected)
			return
		}
		app.mainConnection = conn
		eventClientConnectionAccepted(conn)
	}
}

func processClientMessageTcpPorts(conn net.Conn, message string) {
	if conn != app.mainConnection {
		return
	}
	if tcpPorts := getTcpPortsFromMessage(message); tcpPorts != "" {
		app.potentialTcpPorts = []int{}
		portsString := strings.Split(tcpPorts, ",")
		for _, portString := range portsString {
			port, err := strconv.Atoi(portString)
			if err == nil && checkPort(port) {
				app.potentialTcpPorts = append(app.potentialTcpPorts, port)
			}
		}
	}
}

func closeMainListenerConnection(conn net.Conn, eventFunc func(net.Conn)) {
	serverMutex.Lock()
	defer serverMutex.Unlock()
	if err := conn.Close(); err == nil {
		eventFunc(conn)
	}
}
