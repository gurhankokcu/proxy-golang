package main

import (
	"regexp"
	"strconv"
)

// regexp patterns

var regexpPort = `[0-9]{1,5}`
var regexpServerSecret = `[a-z0-9]{16}`

var regexpConfigServerHost = `^[a-z0-9-\.]{1,250}$`
var regexpConfigCredentials = `^.+$`
var regexpConfigServerSecret = `^` + regexpServerSecret + `$`

var regexpMessageConnection = `^connection=(tcp|udp):(` + regexpPort + `):(` + regexpPort + `)$`
var regexpMessageServerSecret = `^secret=(` + regexpServerSecret + `)$`

var regexpPathTcpPort = `^/admin/tcpports/(` + regexpPort + `)$`
var regexpPathUdpPort = `^/admin/udpports/(` + regexpPort + `)$`

// test regexp

func testRegexp(pattern string, text string) bool {
	return regexp.MustCompile(pattern).Match([]byte(text))
}

func checkPort(port int) bool {
	return port >= 1 && port <= 65535
}

func checkNetwork(network string) bool {
	return network == "tcp" || network == "udp"
}

func checkAppType(appType string) bool {
	return appType == "server" || appType == "client"
}

func checkServerHost(host string) bool {
	return testRegexp(regexpConfigServerHost, host)
}

func checkServerSecret(secret string) bool {
	return testRegexp(regexpConfigServerSecret, secret)
}

func checkCredentials(credentials string) bool {
	return testRegexp(regexpConfigCredentials, credentials)
}

// find regexp

func findRegexp(pattern string, text string) string {
	m := regexp.MustCompile(pattern).FindStringSubmatch(text)
	if m == nil || len(m) < 2 {
		return ""
	}
	return m[1]
}

func getServerSecretFromMessage(message string) string {
	return findRegexp(regexpMessageServerSecret, message)
}

func getConnectionFromMessage(message string) (string, int, int, bool) {
	m := regexp.MustCompile(regexpMessageConnection).FindStringSubmatch(message)
	if m == nil || len(m) < 4 {
		return "", 0, 0, false
	}
	network := m[1]
	if !checkNetwork(network) {
		return "", 0, 0, false
	}
	localPort, err := strconv.Atoi(m[2])
	if err != nil || !checkPort(localPort) {
		return "", 0, 0, false
	}
	proxyPort, err := strconv.Atoi(m[3])
	if err != nil || !checkPort(proxyPort) {
		return "", 0, 0, false
	}
	return network, localPort, proxyPort, true
}

func getTcpPortFromPath(path string) string {
	return findRegexp(regexpPathTcpPort, path)
}

func getUdpPortFromPath(path string) string {
	return findRegexp(regexpPathUdpPort, path)
}
