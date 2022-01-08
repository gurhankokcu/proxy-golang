package main

import (
	"regexp"
)

// regexp patterns

var regexpPort = `[0-9]{1,5}`
var regexpServerSecret = `[a-z0-9]{16}`

var regexpConfigServerHost = `^[a-z0-9-\.]{1,250}$`
var regexpConfigCredentials = `^.+$`
var regexpConfigServerSecret = `^` + regexpServerSecret + `$`

var regexpMessageServerSecret = `^secret=(` + regexpServerSecret + `)$`
var regexpMessageTcpPorts = `^tcpports=((` + regexpPort + `,)*` + regexpPort + `)?$`

var regexpPathTcpPort = `^/admin/tcpports/(` + regexpPort + `)$`
var regexpPathUdpPort = `^/admin/udpports/(` + regexpPort + `)$`

// test regexp

func testRegexp(pattern string, text string) bool {
	return regexp.MustCompile(pattern).Match([]byte(text))
}

func checkPort(port int) bool {
	return port >= 1 && port <= 65535
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

func getTcpPortsFromMessage(message string) string {
	return findRegexp(regexpMessageTcpPorts, message)
}

func getTcpPortFromPath(path string) string {
	return findRegexp(regexpPathTcpPort, path)
}

func getUdpPortFromPath(path string) string {
	return findRegexp(regexpPathUdpPort, path)
}
