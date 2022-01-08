package main

import (
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var serverMutex sync.Mutex

// main listener

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

func acceptConnFromListener(listener net.Listener) (net.Conn, bool) {
	if listener == nil {
		return nil, false
	}
	conn, err := listener.Accept()
	if err != nil {
		logError(err)
		time.Sleep(100 * time.Millisecond)
		return nil, false
	}
	return conn, true
}

func sendMessageToClient(message string) {
	sendMessage(message, eventMessageSentToClient)
}

func requestClientTcpPorts() {
	sendMessageToClient("tcpports")
}

func requestClientConnection(network string, localPort int, proxyPort int) {
	sendMessageToClient("connection=" + network + ":" + strconv.Itoa(localPort) + ":" + strconv.Itoa(proxyPort))
}

func listenMainListenerConnections() {
	for {
		if conn, ok := acceptConnFromListener(app.mainListener); ok {
			eventClientConnectionStarted(conn)
			if app.mainConnection != nil {
				closeMainListenerConnection(conn, eventClientConnectionRejectedAlreadyConnected)
				continue
			}
			checkMainListenerConnection(conn)
			go handleMainListenerConnection(conn)
		} else {
			break
		}
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
			if err == nil && checkPort(port) && indexOfItemInIntSlice(&app.TcpPorts, port) == -1 {
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

// user listener

func openUserListeners() {
	for _, port := range app.TcpPorts {
		if !openUserTcpListener(port) {
			removeTcpPort(port)
		}
	}
	for _, port := range app.UdpPorts {
		if !openUserUdpListener(port) {
			removeUdpPort(port)
		}
	}
}

func openUserTcpListener(port int) bool {
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		logError(err)
		return false
	}
	userTcpListener := UserTcpListener{listener: l}
	app.userTcpListeners[strconv.Itoa(port)] = &userTcpListener
	go listenUserTcpConnections(&userTcpListener)
	return true
}

func listenUserTcpConnections(userTcpListener *UserTcpListener) {
	for {
		if conn, ok := acceptConnFromListener(userTcpListener.listener); ok {
			userTcpConnection := UserTcpConnection{connection: conn}
			userTcpListener.connections = append(userTcpListener.connections, &userTcpConnection)
			openClientTcpListener(&userTcpConnection)
		} else {
			break
		}
	}
}

func openClientTcpListener(userTcpConnection *UserTcpConnection) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		logError(err)
		return
	}
	userTcpConnection.clientListener = l
	localPort := userTcpConnection.connection.LocalAddr().(*net.TCPAddr).Port
	proxyPort := l.Addr().(*net.TCPAddr).Port
	if checkPort(localPort) && checkPort(proxyPort) {
		requestClientConnection("tcp", localPort, proxyPort)
	}
	go listenClientTcpConnections(userTcpConnection)
}

func listenClientTcpConnections(userTcpConnection *UserTcpConnection) {
	for {
		if conn, ok := acceptConnFromListener(userTcpConnection.clientListener); ok {
			userTcpConnection.clientConnection = conn
			go copyIO(userTcpConnection.connection, userTcpConnection.clientConnection)
			go copyIO(userTcpConnection.clientConnection, userTcpConnection.connection)
		} else {
			break
		}
	}
}

func closeUserTcpListener(port int) bool {
	if userTcpListener, ok := app.userTcpListeners[strconv.Itoa(port)]; ok {
		for _, connection := range userTcpListener.connections {
			if connection.clientConnection != nil {
				connection.clientConnection.Close()
			}
			connection.clientConnection = nil
			if connection.clientListener != nil {
				connection.clientListener.Close()
			}
			connection.clientListener = nil
			if connection.connection != nil {
				connection.connection.Close()
			}
			connection.connection = nil
		}
		if userTcpListener.listener != nil {
			userTcpListener.listener.Close()
		}
		userTcpListener.listener = nil
		delete(app.userTcpListeners, strconv.Itoa(port))
		return true
	} else {
		logErrorString("tcp listener not found")
		return false
	}
}

func openUserUdpListener(port int) bool {
	return true
}

func closeUserUdpListener(port int) bool {
	return true
}
