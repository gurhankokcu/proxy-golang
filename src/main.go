package main

var app App

func main() {
	if !readConfig() || !writeConfig() {
		panic("cannot start the application")
	}

	switch app.AppType {
	case "server":
		go openMainListener()
	case "client":
		go openMainConnection()
		go reloadOpenTcpPorts()
	}
	startWebServer()
}
