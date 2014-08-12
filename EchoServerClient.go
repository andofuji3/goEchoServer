package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	RECV_BUF_LEN = 65535
	CONFIG_LINE_SIZE = 1024
)

type Server struct {
	AcceptPort string
}

type Client struct {
	ConnectIP   string
	ConnectPort string
}

type Config struct {
	Server Server `json:"Server"`
	Client Client `json:"Client"`
}

func main() {
	dec := json.NewDecoder(os.Stdin)
	var conf Config
	dec.Decode(&conf)
	fmt.Printf("%+v\n", conf)

	go ServerProc(conf)
	go ClientProc(conf)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(0)
	}()

	for {
		time.Sleep(10 * time.Second)
	}
}

func cleanup() {
	fmt.Println("Finish V21 Test Tool")
}

func ServerProc(conf Config) {
	println("Starting the server")

	listener, err := net.Listen("tcp", "localhost:"+conf.Server.AcceptPort)
	if err != nil {
		println("error listening:", err.Error())
		os.Exit(1)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			println("Error accept:", err.Error())
			return
		}
		go EchoFunc(conn)
	}
}

func EchoFunc(conn net.Conn) {
	for {
		buf := make([]byte, RECV_BUF_LEN)
		n, err := conn.Read(buf)
		if err != nil {
			println("Error reading:", err.Error())
			break
		}
		println("received ", n, " bytes of data")

		println("         |+0 +1 +2 +3 +4 +5 +6 +7 +8 +9 +A +B +C +D +E +F")
		println("---------------------------------------------------------")

		for i := 0; i < n; i++ {
			if i % 0x10 == 0 {
				if i != 0 {
					println("");
				}
				fmt.Printf("%08X |", i / 0x10);
			}
			fmt.Printf("%02X ", buf[i]);
		}

		println("");
		echodata := make([]byte, n)
		copy(echodata, buf)

		//send reply
		_, err = conn.Write(echodata)
		if err != nil {
			println("Error send reply:", err.Error())
			break
		} else {
			println("Reply sent")
		}
	}
	conn.Close()
}

func ClientProc(conf Config) {
	println("Starting the client")
	tcp_addr, err := net.ResolveTCPAddr("tcp", conf.Client.ConnectIP+":"+conf.Client.ConnectPort)
	if err != nil {
		println("error tcp resolve failed", err.Error())
		os.Exit(1)
	}

CONNECT_SERVER:

	tcp_conn, err := net.DialTCP("tcp", nil, tcp_addr)
	if err != nil {
		time.Sleep(1 * time.Second)
		goto CONNECT_SERVER
	}
	println("Connect Server")

	for {
		echo := GetEcho(tcp_conn)
		if echo == "" {
			break
		}
		println("receive success")

		SendEcho(tcp_conn, echo)
		println("echo")
	}

	tcp_conn.Close()
}

func SendEcho(conn *net.TCPConn, msg string) {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		println("Error send request:", err.Error())
	} else {
		println("Request sent")
	}
}

func GetEcho(conn *net.TCPConn) string {
	buf_recever := make([]byte, RECV_BUF_LEN)
	n, err := conn.Read(buf_recever)
	if err != nil {
		println("Error while receive response:", err.Error())
		return ""
	}

	echodata := make([]byte, n)
	copy(echodata, buf_recever)

	return string(echodata)
}
