package main

import (
	"net"
	"time"
)

type Event struct {
	text string
}

func NewEvent(text string) *Event {
	return &Event{text: time.Now().Format(time.DateTime) + ": " + text}
}

func (e *Event) append(t string) *Event {
	if t != "" {
		e.text += " | " + t
	}
	return e
}

func (e *Event) from(remote, local net.Addr) *Event {
	if remote != nil {
		e.text += " | " + remote.String() + " (" + remote.Network() +
			") -> " + local.String() + " (" + local.Network() + ")"
	}
	return e
}

func (e *Event) to(local, remote net.Addr) *Event {
	if remote != nil {
		e.text += " | " + local.String() + " (" + local.Network() +
			") -> " + remote.String() + " (" + remote.Network() + ")"
	}
	return e
}

func (e Event) broadcast() {
	addEvent(e.text)
	for _, l := range app.adminListeners {
		eventResponse(l, e.text)
	}
}

// Generic events

func eventLog(text string) {
	NewEvent(text).broadcast()
}

func eventError(text string) {
	NewEvent("Error: " + text).broadcast()
}

// Server events

func eventClientConnectionStarted(conn net.Conn) {
	NewEvent("Client connection started").from(conn.RemoteAddr(), conn.LocalAddr()).broadcast()
}

func eventClientConnectionAccepted(conn net.Conn) {
	NewEvent("Client connection accepted").from(conn.RemoteAddr(), conn.LocalAddr()).broadcast()
}

func eventClientConnectionRejectedTimeout(conn net.Conn) {
	NewEvent("Client connection rejected, timed out").from(conn.RemoteAddr(), conn.LocalAddr()).broadcast()
}

func eventClientConnectionRejectedInvalidSecret(conn net.Conn) {
	NewEvent("Client connection rejected, invalid secret").from(conn.RemoteAddr(), conn.LocalAddr()).broadcast()
}

func eventClientConnectionRejectedAlreadyConnected(conn net.Conn) {
	NewEvent("Client connection rejected, client already connected").from(conn.RemoteAddr(), conn.LocalAddr()).broadcast()
}

func eventClientConnectionEnded(conn net.Conn) {
	NewEvent("Client connection ended").from(conn.RemoteAddr(), conn.LocalAddr()).broadcast()
}

func eventMessageReceivedFromClient(conn net.Conn, message string) {
	NewEvent("Message received from client").append(message).from(conn.RemoteAddr(), conn.LocalAddr()).broadcast()
}

func eventMessageSentToClient(conn net.Conn, message string) {
	NewEvent("Message sent to client").append(message).to(conn.LocalAddr(), conn.RemoteAddr()).broadcast()
}

// Client events

func eventServerConnectionStarted(conn net.Conn) {
	NewEvent("Server connection started").from(conn.RemoteAddr(), conn.LocalAddr()).broadcast()
}

func eventServerConnectionEnded(conn net.Conn) {
	NewEvent("Server connection ended").from(conn.RemoteAddr(), conn.LocalAddr()).broadcast()
}

func eventMessageReceivedFromServer(conn net.Conn, message string) {
	NewEvent("Message received from server").append(message).from(conn.RemoteAddr(), conn.LocalAddr()).broadcast()
}

func eventMessageSentToServer(conn net.Conn, message string) {
	NewEvent("Message sent to server").append(message).to(conn.LocalAddr(), conn.RemoteAddr()).broadcast()
}

// Proxy events

func eventProxyConnection(local, proxy net.Conn) {
	NewEvent("Proxy connection").from(local.RemoteAddr(), local.LocalAddr()).to(proxy.LocalAddr(), proxy.RemoteAddr()).broadcast()
}

func eventProxyUdpConnection(local, proxy *net.UDPConn) {
	NewEvent("Proxy connection").from(local.RemoteAddr(), local.LocalAddr()).to(proxy.LocalAddr(), proxy.RemoteAddr()).broadcast()
}
