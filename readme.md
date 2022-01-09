# Proxy

Proxy application is to make a computer on a private network publicly accessable.

Enabling outside access to an internal computer on a private network usually requires changing router/modem settings and setting up NAT (network address translation) or port forwarding. This configuration forwards the requests made to your modem to the right computer on the internal network.

This application allows an internal computer to be accessed from outside world without any changes to the modem or router. The application should be run as a pair, a server and a client. Server should be run on a public computer and client should be run on the computer on a private network. When the client is connected to the server, the ports can be enabled for public access individually.

## Requirements

* [Go](https://go.dev/)

## Installation

Install Go using the link above

Clone the GitHub repository

```bash
$ cd /path/to/destination/
$ git clone https://github.com/gurhankokcu/proxy-golang.git
```

Run build.sh to create executables for Linux and Mac

```bash
$ cd proxy-golang
$ chmod u+x build.sh
$ build.sh
```

Executable file, configuration file and html files can be found by the platform (linux/mac) in the `bin` folder.

## Configuration

Proxy application can be run as a server or a client according to the configuration. After building the project (running the `build.sh` file), `config.js` will be copied into the `bin/linux` and `bin/mac` folders.

The configuration file (`config.js`) should be changed to define the type of the application. `appType` property defines either the application is a `server` or a `client`.

Sample server configuration:
```json
{
    "appType": "server",
    "serverPort": 9876,
    "serverSecret": "my-server-secret",
    "adminPort": 86,
    "adminUser": "admin",
    "adminPass": "password",
    "tcpPorts": [],
    "udpPorts": []
}
```

Sample client configuration:
```json
{
    "appType": "client",
    "serverHost": "127.0.0.1",
    "serverPort": 9876,
    "serverSecret": "my-server-secret",
    "adminPort": 91,
    "adminUser": "admin",
    "adminPass": "password",
    "tcpPorts": [],
    "udpPorts": []
}
```

## Usage

Run executable golang application

```bash
$ cd /path/to/destination/
$ ./proxy
```

Configuration can be overridden by the arguments. None of these arguments are required, configuration file will be used for the missing arguments.

Executing with server configuration:
```bash
$ cd /path/to/destination/
$ ./proxy --app-type=server --server-port=9876 --server-secret=my-server-secret --admin-port=86 --admin-user=admin --admin-pass=password
```

Executing with client configuration:
```bash
$ cd /path/to/destination/
$ ./proxy --app-type=client --server-host=127.0.0.1 --server-port=9876 --server-secret=my-server-secret --admin-port=91 --admin-user=admin --admin-pass=password
```

After running the server and the client, the client should connect to the server using server host and server port configuration.

Admin web interface helps to change all of the configuration, reconnect client to the server and enable/disable ports.

Server admin web interface
```http
http://server-address:86/admin/
```

Client admin web interface
```http
http://localhost:91/admin/
```

## License

[ISC](https://choosealicense.com/licenses/isc/)
