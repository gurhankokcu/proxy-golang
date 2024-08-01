package main

import (
	"net"
)

type Event struct {
	text string
}

func NewEvent(text string) *Event {
	return &Event{text}
}

func (e *Event) append(t string) *Event {
	if t != "" {
		e.text += " | " + t
	}
	return e
}

func (e *Event) from(conn net.Conn) *Event {
	if conn != nil {
		e.text += " | " + conn.RemoteAddr().String() + " (" + conn.RemoteAddr().Network() +
			") -> " + conn.LocalAddr().String() + " (" + conn.LocalAddr().Network() + ")"
	}
	return e
}

func (e *Event) to(conn net.Conn) *Event {
	if conn != nil {
		e.text += " | " + conn.LocalAddr().String() + " (" + conn.LocalAddr().Network() +
			") -> " + conn.RemoteAddr().String() + " (" + conn.RemoteAddr().Network() + ")"
	}
	return e
}

func (e Event) broadcast() {
	logInfo(e.text)
	for _, l := range app.adminListeners {
		eventResponse(l, e.text)
	}
}

// Server events

func eventClientConnectionStarted(conn net.Conn) {
	NewEvent("Client connection started").from(conn).broadcast()
}

func eventClientConnectionAccepted(conn net.Conn) {
	NewEvent("Client connection accepted").from(conn).broadcast()
}

func eventClientConnectionRejectedTimeout(conn net.Conn) {
	NewEvent("Client connection rejected, timed out").from(conn).broadcast()
}

func eventClientConnectionRejectedInvalidSecret(conn net.Conn) {
	NewEvent("Client connection rejected, invalid secret").from(conn).broadcast()
}

func eventClientConnectionRejectedAlreadyConnected(conn net.Conn) {
	NewEvent("Client connection rejected, client already connected").from(conn).broadcast()
}

func eventClientConnectionEnded(conn net.Conn) {
	NewEvent("Client connection ended").from(conn).broadcast()
}

func eventMessageReceivedFromClient(conn net.Conn, message string) {
	NewEvent("Message received from client").append(message).from(conn).broadcast()
}

func eventMessageSentToClient(conn net.Conn, message string) {
	NewEvent("Message sent to client").append(message).to(conn).broadcast()
}

// Client events

func eventServerConnectionStarted(conn net.Conn) {
	NewEvent("Server connection started").from(conn).broadcast()
}

func eventServerConnectionEnded(conn net.Conn) {
	NewEvent("Server connection ended").from(conn).broadcast()
}

func eventMessageReceivedFromServer(conn net.Conn, message string) {
	NewEvent("Message received from server").append(message).from(conn).broadcast()
}

func eventMessageSentToServer(conn net.Conn, message string) {
	NewEvent("Message sent to server").append(message).to(conn).broadcast()
}

// Proxy events

func eventProxyConnection(local, proxy net.Conn) {
	NewEvent("Proxy connection").from(local).to(proxy).broadcast()
}

// func eventMessageSendingError(conn net.Conn, message string) {
// 	NewEvent("Error while sending a message").append(message).to(conn).broadcast()
// }
