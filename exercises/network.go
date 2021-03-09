package main

import (
	"bufio"
	"fmt"
	"net"
	"syscall"
	"time"
)

// in golang, you dont create and then bind a socket. You create and bind in a single function call.
// So there is no time to set socket options in between!
// For UDP broadcast: But look at the go-networking module in TTK4145, he has implemented this for us
// not for windows

// Socket options: SET THESE BEFORE BINDING/CONNECTING/ETC
// * SO_BROADCAST to enable UDP broadcast
// * SO_REUSEADDR to use an adress that is already in use (Messages should be recieved by all programs that use that port)
//  * Can broadcast between programs, not only networks, useful for developing at home
// * TCP_NODELAY

const UDP_PACKET_SIZE = 64
const TCP_PACKET_SIZE = 128

func print_broadcast_on(networkType string, port string) {
	addr, err := net.ResolveUDPAddr(networkType, ":"+port)
	if err != nil {
		fmt.Println(err)
	}
	conn, err := net.ListenUDP(networkType, addr)
	if err != nil {
		fmt.Println(err)
	}
	buf := make([]byte, UDP_PACKET_SIZE)
	for {
		// n, remoteAdress, error
		n, _, _ := conn.ReadFromUDP(buf)
		fmt.Println(string(buf[0:n]))
		time.Sleep(1000)
	}
}

func send_message_udp(networkType, ipAddr, port string, message string) {
	remoteAddr, err := net.ResolveUDPAddr(networkType, ipAddr+":"+port)
	if err != nil {
		fmt.Println(err)
	}

	conn, err := net.DialUDP(networkType, nil, remoteAddr)
	if err != nil {
		fmt.Println(err)
	}

	defer conn.Close()

	byteMessage := []byte(message)
	_, err = conn.Write(byteMessage)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Message sent: " + message)

	return
}

func recieve_message_udp(networkType, port string, finished chan bool) {
	addr, err := net.ResolveUDPAddr(networkType, ":"+port)
	if err != nil {
		fmt.Println(err)
	}
	conn, err := net.ListenUDP(networkType, addr)
	if err != nil {
		fmt.Println(err)
	}

	buf := make([]byte, UDP_PACKET_SIZE)
	for {
		n, _ := conn.Read(buf)
		fmt.Println(string(buf[0:n]))
		time.Sleep(1000 * time.Millisecond)
	}

	finished <- true

	return
}

func connect_to_server_tcp(ipaddr, port string) net.Conn {
	raddr, err := net.ResolveTCPAddr("tcp", ipaddr+":"+port)
	if err != nil {
		fmt.Println(err)
	}

	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		fmt.Println(err)
	}

	return conn
}

func start_reading_msgs_from(conn net.Conn, packetSize int) {
	for {
		message, err := bufio.NewReader(conn).ReadString('\000')
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(message)
		time.Sleep(1000 * time.Millisecond)
	}
}

func send_message_tcp(conn net.Conn, message string) {
	byteMessage := []byte(message)

	_, err := conn.Write(byteMessage)
	if err != nil {
		fmt.Println(err)
	}
	return
}

func start_accepting_tcp_connections(port string) net.Conn {
	laddr, err := net.ResolveTCPAddr("tcp", ":"+port)
	if err != nil {
		fmt.Println(err)
	}
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		fmt.Println(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
		}
		go handle_connection(conn)
	}
}

func handle_connection(conn net.Conn) {
	go start_reading_msgs_from(conn, 128)
	time.Sleep(1000 * time.Millisecond)
	go send_message_tcp(conn, "The connection was accepted!\000")
}

func main() {
	s, _ := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)

	finished := make(chan bool, 2)

	// go print_broadcast_on("udp", "30000")

	message := "Hej server.\000"
	go recieve_message_udp("udp", "20001", finished)
	go send_message_udp("udp", "192.168.0.151", "20000", message)

	conn := connect_to_server_tcp("192.168.0.151", "33546")
	go start_reading_msgs_from(conn, 128)
	time.Sleep(1000 * time.Millisecond)
	send_message_tcp(conn, "Hej")
	send_message_tcp(conn, "Eric")
	send_message_tcp(conn, "Martin!!\000")

	go start_accepting_tcp_connections("33333")
	send_message_tcp(conn, "Connect to: 192.168.0.151:33333\000")

	for {
		select {
		case <-finished:
			select {
			case <-finished:
				fmt.Println("End of script.")
			}
		}
	}
}
