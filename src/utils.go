package main

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strconv"
)

func logInfo(text string) {
	fmt.Println(text)
}

func logError(err error) {
	logErrorString(err.Error())
}

func logErrorString(err string) {
	fmt.Println(err)
	eventError(err)
}

func indexOfItemInIntSlice(slice *[]int, item int) int {
	for index, value := range *slice {
		if value == item {
			return index
		}
	}
	return -1
}

func getIntValue(val1, val2, val3 int, check func(int) bool) int {
	if check(val1) {
		return val1
	}
	if check(val2) {
		return val2
	}
	if check(val3) {
		return val3
	}
	return 0
}

func getStringValue(val1, val2, val3 string, check func(string) bool) string {
	if check(val1) {
		return val1
	}
	if check(val2) {
		return val2
	}
	if check(val3) {
		return val3
	}
	return ""
}

func getIntSliceValue(val1, val2, val3 []int, check func(int) bool) []int {
	var items []int
	if len(val1) > 0 {
		items = val1
	}
	if len(val2) > 0 {
		items = val2
	}
	if len(val3) > 0 {
		items = val3
	}
	var result = []int{}
	for _, item := range items {
		if check(item) && indexOfItemInIntSlice(&result, item) == -1 {
			result = append(result, item)
		}
	}
	return result
}

func randomString(n int) string {
	var chars = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = chars[rand.Intn(len(chars))]
	}
	return string(s)
}

func isTcpPortOpen(port int) bool {
	l, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if l != nil {
		l.Close()
	}
	return err == nil
}

func getPublicIP() string {
	resp, err := http.Get("https://api.ipify.org?format=text")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(ip)
}
