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
		// go closeUserListeners()
		time.Sleep(1 * time.Second)
		go openUserListeners()
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

// user listener

func openUserListeners() {
	for {
		if app.IsConnected() {
			break
		}
		time.Sleep(1 * time.Second)
	}
	for _, port := range app.TcpPorts {
		go openUserTcpListener(port)
	}
	for _, port := range app.UdpPorts {
		go openClientUdpConnection(port)
	}
}

// func closeUserListeners() {
// 	for _, port := range app.TcpPorts {
// 		go closeUserTcpListener(port)
// 	}
// 	for _, port := range app.UdpPorts {
// 		go closeClientUdpConnection(port)
// 	}
// }

// TCP

func openUserTcpListener(port int) {
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		logError(err)
		removeTcpPort(port)
		return
	}
	userTcpListener := UserTcpListener{listener: l}
	app.userTcpListeners[strconv.Itoa(port)] = &userTcpListener
	listenUserTcpConnections(&userTcpListener)
}

func listenUserTcpConnections(userTcpListener *UserTcpListener) {
	for {
		if conn, ok := acceptConnFromListener(userTcpListener.listener); ok {
			userTcpConnection := UserTcpConnection{connection: conn}
			userTcpListener.connections = append(userTcpListener.connections, &userTcpConnection)
			go openClientTcpListener(&userTcpConnection)
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
	listenClientTcpConnections(userTcpConnection)
}

func listenClientTcpConnections(userTcpConnection *UserTcpConnection) {
	for {
		if conn, ok := acceptConnFromListener(userTcpConnection.clientListener); ok {
			userTcpConnection.clientConnection = conn
			eventProxyConnection(userTcpConnection.connection, userTcpConnection.clientConnection)
			go copyIO(userTcpConnection.connection, userTcpConnection.clientConnection)
		} else {
			break
		}
	}
}

func closeUserTcpListener(port int) {
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
	} else {
		logErrorString("tcp listener not found")
	}
}

// UDP

func openClientUdpConnection(port int) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 0})
	if err != nil {
		logError(err)
		return
	}
	clientUdpConnection := ClientUdpConnection{connection: conn}
	app.clientUdpConnections[strconv.Itoa(port)] = &clientUdpConnection

	proxyPort := conn.LocalAddr().(*net.UDPAddr).Port

	if checkPort(port) && checkPort(proxyPort) {
		requestClientConnection("udp", port, proxyPort)
	}

	receiveClientUdpConnection(&clientUdpConnection)
	time.Sleep(100 * time.Millisecond)
	listenUserUdpConnections(&clientUdpConnection, port)
}

func receiveClientUdpConnection(clientUdpConnection *ClientUdpConnection) {
	buffer := make([]byte, 4096)
	n, clientAddr, err := clientUdpConnection.connection.ReadFromUDP(buffer)
	if err != nil {
		logError(err)
		return
	}
	clientUdpConnection.remoteAddr = clientAddr
	logInfo("client udp connection: " + string(buffer[:n]) + " " + clientAddr.String())
}

func listenUserUdpConnections(clientUdpConnection *ClientUdpConnection, port int) {
	logInfo("start listening udp connections on port " + strconv.Itoa(port))
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: port})
	if err != nil {
		logError(err)
		return
	}
	logInfo("udp connection received")

	userUdpConnection := UserUdpConnection{connection: conn}
	clientUdpConnection.userConnections = append(clientUdpConnection.userConnections, &userUdpConnection)

	eventProxyUdpConnection(userUdpConnection.connection, clientUdpConnection.connection)
	udpCopyIO(&userUdpConnection, clientUdpConnection)
}

func closeClientUdpConnection(port int) {
	if clientUdpConnection, ok := app.clientUdpConnections[strconv.Itoa(port)]; ok {
		for _, connection := range clientUdpConnection.userConnections {
			if connection.connection != nil {
				connection.connection.Close()
			}
			connection.connection = nil
			connection.remoteAddr = nil
		}
		clientUdpConnection.userConnections = nil
		if clientUdpConnection.connection != nil {
			clientUdpConnection.connection.Close()
		}
		clientUdpConnection.connection = nil
		clientUdpConnection.remoteAddr = nil
		delete(app.clientUdpConnections, strconv.Itoa(port))
	} else {
		logErrorString("udp listener not found")
	}
}
