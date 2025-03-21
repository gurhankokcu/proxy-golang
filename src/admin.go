package main

import (
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"
)

func startWebServer() {
	http.HandleFunc("/admin/server/", handleAdminServer)
	http.HandleFunc("/admin/admin/", handleAdminAdmin)
	http.HandleFunc("/admin/tcpports/", handleAdminTcpPorts)
	http.HandleFunc("/admin/udpports/", handleAdminUdpPorts)
	http.HandleFunc("/admin/reconnect/", handleAdminReconnect)
	http.HandleFunc("/admin/requestports/", handleAdminRequestPorts)
	http.HandleFunc("/admin/events/", handleAdminEvents)
	http.HandleFunc("/admin/", handleAdmin)
	http.HandleFunc("/", handleMain)
	http.ListenAndServe(":"+strconv.Itoa(app.AdminPort), nil)
}

// TODO: encrypt username and password
func basicAuth(w http.ResponseWriter, req *http.Request) bool {
	username, password, ok := req.BasicAuth()
	if !ok || subtle.ConstantTimeCompare([]byte(username), []byte(app.AdminUser)) != 1 || subtle.ConstantTimeCompare([]byte(password), []byte(app.AdminPass)) != 1 {
		w.Header().Set("WWW-Authenticate", `Basic realm="Please enter admin username and password"`)
		w.WriteHeader(http.StatusUnauthorized)
		template.Must(template.ParseFiles("unauthorized.html", "layout.html")).Execute(w, &app)
		return false
	}
	return true
}

func errResponse(w http.ResponseWriter, err string) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err))
}

func okResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func eventResponse(w http.ResponseWriter, text string) {
	if f, ok := w.(http.Flusher); ok {
		w.Write([]byte(text + "\r\n"))
		f.Flush()
	}
}

func handleMain(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("It works!"))
}

func handleAdmin(w http.ResponseWriter, req *http.Request) {
	if !basicAuth(w, req) {
		return
	}

	template.Must(template.ParseFiles("admin.html", "layout.html")).Execute(w, &app)
}

func handleAdminServer(w http.ResponseWriter, req *http.Request) {
	if !basicAuth(w, req) {
		return
	}

	switch req.Method {
	case http.MethodPut:
		byteValue, err := io.ReadAll(req.Body)
		if err != nil {
			errResponse(w, err.Error())
			return
		}
		var a App
		err = json.Unmarshal(byteValue, &a)
		if err != nil {
			errResponse(w, err.Error())
			return
		}
		existingServerHost := app.ServerHost
		existingServerPort := app.ServerPort
		existingServerSecret := app.ServerSecret
		if !updateServer(a.ServerHost, a.ServerPort, a.ServerSecret) {
			app.ServerHost = existingServerHost
			app.ServerPort = existingServerPort
			app.ServerSecret = existingServerSecret
			errResponse(w, "An error occurred while saving changes")
			return
		}
		switch app.AppType {
		case "server":
			closeMainListener()
			go openMainListener()
		case "client":
			closeMainConnection()
			go openMainConnection()
		}
		okResponse(w)
	}
}

func handleAdminAdmin(w http.ResponseWriter, req *http.Request) {
	if !basicAuth(w, req) {
		return
	}

	switch req.Method {
	case http.MethodPut:
		byteValue, err := io.ReadAll(req.Body)
		if err != nil {
			errResponse(w, err.Error())
			return
		}
		var a App
		err = json.Unmarshal(byteValue, &a)
		if err != nil {
			errResponse(w, err.Error())
			return
		}
		existingPort := app.AdminPort
		existingUser := app.AdminUser
		existingPass := app.AdminPass
		if !updateAdmin(a.AdminPort, a.AdminUser, a.AdminPass) {
			app.AdminPort = existingPort
			app.AdminUser = existingUser
			app.AdminPass = existingPass

			errResponse(w, "An error occurred while saving changes")
			return
		}
		okResponse(w)
		time.AfterFunc(time.Second, func() {
			os.Exit(1)
		})
	}
}

func handleAdminPortsFunc(getPort func(string) string, openListener, closeListener func(int), addPort, removePort func(int) bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if !basicAuth(w, req) {
			return
		}

		switch req.Method {
		case http.MethodPost:
			byteValue, err := io.ReadAll(req.Body)
			if err != nil {
				errResponse(w, err.Error())
				return
			}
			body := string(byteValue)
			port, err := strconv.Atoi(body)
			if err != nil {
				errResponse(w, err.Error())
				return
			}
			if !addPort(port) {
				errResponse(w, "Invalid or duplicate port "+strconv.Itoa(port))
				return
			}
			if app.AppType == "server" {
				go openListener(port)
			}
			okResponse(w)
		case http.MethodDelete:
			port, err := strconv.Atoi(getPort(req.URL.Path))
			if err != nil {
				errResponse(w, err.Error())
				return
			}
			if !removePort(port) {
				errResponse(w, "Port "+strconv.Itoa(port)+" not found")
				return
			}
			if app.AppType == "server" {
				go closeListener(port)
			}
			okResponse(w)
		}
	}
}

func handleAdminTcpPorts(w http.ResponseWriter, req *http.Request) {
	handleAdminPortsFunc(getTcpPortFromPath, openUserTcpListener, closeUserTcpListener, addTcpPort, removeTcpPort)(w, req)
}

func handleAdminUdpPorts(w http.ResponseWriter, req *http.Request) {
	handleAdminPortsFunc(getUdpPortFromPath, openClientUdpConnection, closeClientUdpConnection, addUdpPort, removeUdpPort)(w, req)
}

func handleAdminReconnect(w http.ResponseWriter, req *http.Request) {
	if !basicAuth(w, req) {
		return
	}
	switch req.Method {
	case http.MethodGet:
		switch app.AppType {
		case "server":
		case "client":
			closeMainConnection()
			go openMainConnection()
		}
		okResponse(w)
	}
}

func handleAdminRequestPorts(w http.ResponseWriter, req *http.Request) {
	if !basicAuth(w, req) {
		return
	}
	switch req.Method {
	case http.MethodGet:
		switch app.AppType {
		case "server":
			requestClientTcpPorts()
		case "client":
		}
		okResponse(w)
	}
}

func handleAdminEvents(w http.ResponseWriter, req *http.Request) {
	if !basicAuth(w, req) {
		return
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	app.adminListeners = append(app.adminListeners, w)

	ctx := req.Context()
	<-ctx.Done()
	index := -1
	for i, l := range app.adminListeners {
		if l == w {
			index = i
			break
		}
	}
	if index != -1 {
		app.adminListeners = append(app.adminListeners[:index], app.adminListeners[index+1:]...)
	}
}
