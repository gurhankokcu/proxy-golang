package main

import (
	"encoding/json"
	"flag"
	"os"
	"sort"
)

func getAppFlags() *App {
	var app App
	flag.StringVar(&app.AppType, "app-type", "", "application type")
	flag.StringVar(&app.ServerHost, "server-host", "", "server host")
	flag.IntVar(&app.ServerPort, "server-port", 0, "server port")
	flag.StringVar(&app.ServerSecret, "server-secret", "", "server secret")
	flag.IntVar(&app.AdminPort, "admin-port", 0, "admin page port")
	flag.StringVar(&app.AdminUser, "admin-user", "", "admin page username")
	flag.StringVar(&app.AdminPass, "admin-pass", "", "admin page password")
	flag.Parse()
	return &app
}

func getAppConfig() *App {
	var app App
	byteValue, err := os.ReadFile("config.json")
	if err != nil {
		logError(err)
		return &app
	}
	err = json.Unmarshal(byteValue, &app)
	if err != nil {
		logError(err)
		return &app
	}
	return &app
}

func getAppDefault() *App {
	return &App{
		AppType:      "server",
		ServerHost:   "127.0.0.1",
		ServerPort:   9876,
		ServerSecret: randomString(16),
		AdminPort:    86,
		AdminUser:    "admin",
		AdminPass:    "password",
		TcpPorts:     make([]int, 0),
		UdpPorts:     make([]int, 0),
		Events:       make([]string, 0),
	}
}

func readConfig() bool {
	appFlags := getAppFlags()
	appConfig := getAppConfig()
	appDefault := getAppDefault()

	app.AppType = getStringValue(appFlags.AppType, appConfig.AppType, appDefault.AppType, checkAppType)
	app.ServerHost = getStringValue(appFlags.ServerHost, appConfig.ServerHost, appDefault.ServerHost, checkServerHost)
	app.ServerPort = getIntValue(appFlags.ServerPort, appConfig.ServerPort, appDefault.ServerPort, checkPort)
	app.ServerSecret = getStringValue(appFlags.ServerSecret, appConfig.ServerSecret, appDefault.ServerSecret, checkServerSecret)
	app.AdminPort = getIntValue(appFlags.AdminPort, appConfig.AdminPort, appDefault.AdminPort, checkPort)
	app.AdminUser = getStringValue(appFlags.AdminUser, appConfig.AdminUser, appDefault.AdminUser, checkCredentials)
	app.AdminPass = getStringValue(appFlags.AdminPass, appConfig.AdminPass, appDefault.AdminPass, checkCredentials)
	app.TcpPorts = getIntSliceValue(appFlags.TcpPorts, appConfig.TcpPorts, appDefault.TcpPorts, checkPort)
	app.UdpPorts = getIntSliceValue(appFlags.UdpPorts, appConfig.UdpPorts, appDefault.UdpPorts, checkPort)
	app.Events = appConfig.Events
	if app.Events == nil {
		app.Events = appDefault.Events
	}

	app.Ip = getPublicIP()
	app.userTcpListeners = make(map[string]*UserTcpListener)
	app.clientUdpConnections = make(map[string]*ClientUdpConnection)

	return true
}

func writeConfig() bool {
	byteValue, err := json.MarshalIndent(&app, "", "    ")
	if err != nil {
		return false
	}
	err = os.WriteFile("config.json", byteValue, 0644)
	return err == nil
}

func updateServer(serverHost string, serverPort int, serverSecret string) bool {
	if checkServerHost(serverHost) {
		app.ServerHost = serverHost
	}
	if checkPort(serverPort) {
		app.ServerPort = serverPort
	}
	if checkServerSecret(serverSecret) {
		app.ServerSecret = serverSecret
	}
	return writeConfig()
}

func updateAdmin(port int, user string, pass string) bool {
	if checkPort(port) {
		app.AdminPort = port
	}
	if checkCredentials(user) {
		app.AdminUser = user
	}
	if checkCredentials(pass) {
		app.AdminPass = pass
	}
	return writeConfig()
}

func addTcpPort(port int) bool {
	if !checkPort(port) {
		return false
	}
	if indexOfItemInIntSlice(&app.TcpPorts, port) != -1 {
		return false
	}
	app.TcpPorts = append(app.TcpPorts, port)
	sort.Slice(app.TcpPorts, func(i, j int) bool {
		return app.TcpPorts[i] < app.TcpPorts[j]
	})
	return writeConfig()
}

func removeTcpPort(port int) bool {
	index := indexOfItemInIntSlice(&app.TcpPorts, port)
	if index == -1 {
		return false
	}
	app.TcpPorts = append(app.TcpPorts[:index], app.TcpPorts[index+1:]...)
	return writeConfig()
}

func addUdpPort(port int) bool {
	if !checkPort(port) {
		return false
	}
	if indexOfItemInIntSlice(&app.UdpPorts, port) != -1 {
		return false
	}
	app.UdpPorts = append(app.UdpPorts, port)
	sort.Slice(app.UdpPorts, func(i, j int) bool {
		return app.UdpPorts[i] < app.UdpPorts[j]
	})
	return writeConfig()
}

func removeUdpPort(port int) bool {
	index := indexOfItemInIntSlice(&app.UdpPorts, port)
	if index == -1 {
		return false
	}
	app.UdpPorts = append(app.UdpPorts[:index], app.UdpPorts[index+1:]...)
	return writeConfig()
}

func addEvent(message string) bool {
	app.Events = append(app.Events, message)
	return writeConfig()
}
