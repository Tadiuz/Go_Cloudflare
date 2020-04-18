/*
Created by Rodrigo Estrada Alducin on April 4th 2020 for the Cloudflare internship 2020

######Github: https://github.com/Tadiuz/Go_Cloudflare.git

I used some resources to get inspired, pricipally the The go Playground example of how to use the library ICMP in go
Link: https://play.golang.org/p/WPmFYG51KQq

The tool runs on linux and it has to be run with root privileges

This is my first time programing in Golang and also the first time using the ICM protocol so I
used some resoucers to know more about it.

To understant how ICMP works i used this resouce: http://www.tcpipguide.com/free/t_ICMPv4EchoRequestandEchoReplyMessages-2.htm

And also the documentation of the ICMP library for golang: https://godoc.org/golang.org/x/net/icmp

I tested the program on ubuntu trough irtual box, so the ipv6 protocol wasn't worlong at all,
i needed to install miredo to simulate and ipv6 tunnel troight ipv4 (This is just to test it on a virtual machine)
$ sudo apt-get install miredo


Usage: [options] host [Host Name/ Address]
Options:
  -4    use IPv4 protocol
  -6    use IPv6 protocol
  -TTL float
        Time for TTL response in seconds (default 10)
  -mssm string
        Define your own message to send (default "Hello Cloudflare")

Example:

sudo go run Go_ping_Rodrigo.go -4 -TTL=5 -mssm= "Hello World" google.com

Im'not a software engineer so I cant say that the logic implement here was the best one, but I did my best haha :D

*/

package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

var (
	/*
		In the context of servers, 0.0.0.0 means all IPv4 addresses on the local machine.
		If a host has two IP addresses, 192.168.1.1 and 10.1.2.1,
		and a server running on the host listens on 0.0.0.0, it will be reachable at both of
		those IPs.
		The listen addres is to test that everything is working correctly with sending ICP packages
	*/
	ListenAddr   = "0.0.0.0"
	ListenAddrv6 = "::" //Default for ipv6

	send     float64
	received float64
	loss     float64
)

func getResponse(addr string, mssm string, timer float64) (*net.IPAddr, time.Duration, string, error) {

	//We're opening the channel to receive packages, with the 0.0.0.0 ip direction that mean from every ip
	listener, err := icmp.ListenPacket("ip4:icmp", ListenAddr)

	defer listener.Close() //With defer we close the connection at the end of the function

	//fmt.Println("listener VAlue: ", listener, "Error VAlue:", err)

	if err != nil {

		return nil, 0, "", err
	}

	ipAddrss, err := net.ResolveIPAddr("ip4", addr)

	//WE use this funtion to get the real ip addres in case we put the host and not the ip directly
	//fmt.Println("ipAddrss VAlue: ", ipAddrss, "Error VAlue:", err)

	if err != nil {
		fmt.Println("Invalid Host, please try again...")
		panic(err)

		return nil, 0, "", err
	}

	//icmp.Message its an struct
	/*
		type Message struct {
			Type     Type        // type, either ipv4.ICMPType or ipv6.ICMPType
			Code     int         // code
			Checksum int         // checksum
			Body     MessageBody // body
	}*/

	// The type ipv4.ICMPTypeEcho is t send a request , Code: Not used for Echo and Echo Reply messages; set to 0,
	//For echo request de body is made of a Identifiers, Sequence NUmber and Payload
	//FOr the identifier is common to use the Id of the current proccess and use logic gate with en hex value
	//IN that way everytime that we restart the porcess itll be a differetn identifier

	icpmToSend := icmp.Message{Type: ipv4.ICMPTypeEcho, Code: 0, Body: &icmp.Echo{ID: os.Getpid() & 0xffff, Seq: 1, Data: []byte(mssm)}}
	//fmt.Println("first id value", os.Getpid()&0xffff)

	//Marshal return the binary encoding of the Message
	binaryEncoding, err := icpmToSend.Marshal(nil)
	//fmt.Println("binaryEncoding value: ", (binaryEncoding), " Error value: ", err)

	if err != nil {

		return ipAddrss, 0, "", err
	}

	//Here well send the request
	//start := time.Now()

	/*
		WriteTo writes the ICMP message b to dst. The provided dst must be net.UDPAddr when c is a non-privileged datagram-oriented ICMP endpoint. Otherwise it must be net.IPAddr.
		from : https://godoc.org/golang.org/x/net/icmp#PacketConn
	*/

	start := time.Now()
	n, err := listener.WriteTo(binaryEncoding, ipAddrss)
	//fmt.Println("N value", n, " Error value: ", err)

	if err != nil {
		return ipAddrss, 0, "", err
	} else if n != len(binaryEncoding) {
		return ipAddrss, 0, "", fmt.Errorf("got %v; want %v", n, len(binaryEncoding))
	}

	// Wait for a reply
	reply := make([]byte, 1500)

	//Set the maximum time to wait for the response, in this case well wait 10 seconds
	err = listener.SetReadDeadline(time.Now().Add(time.Duration(timer) * time.Second))
	//fmt.Println(" Error value: ", err)

	if err != nil {

		return ipAddrss, 0, "", err
	}

	n, peer, err := listener.ReadFrom(reply)

	//fmt.Println("n value", n, " Error value: ", err, "Peer Value", peer)
	if err != nil {
		log.Println(".....time exceeded.....")
		return ipAddrss, 0, "", err
	}
	duration := time.Since(start)

	//fmt.Println("duration value", duration)

	rm, err := icmp.ParseMessage(1, reply[:n])
	v := string(reply[8:n])

	//fmt.Println("rm value", v, "Error value: ", err, "reply: ")

	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		return ipAddrss, duration, v, nil
	default:
		return ipAddrss, 0, "", fmt.Errorf("got %+v from %v; want echo reply", rm, peer)
	}
}

func getResponsev6(addr string, mssm string, timer float64) (*net.IPAddr, time.Duration, string, error) {

	//We're opening the channel to receive packages, with the 0.0.0.0 ip direction that mean from every ip
	listener, err := icmp.ListenPacket("ip6:ipv6-icmp", ListenAddrv6)

	defer listener.Close() //With defer we close the connection at the end of the function

	//fmt.Println("listener VAlue: ", listener, "Error VAlue:", err)

	if err != nil {

		return nil, 0, "", err
	}

	ipAddrss, err := net.ResolveIPAddr("ip6", addr)

	//WE use this funtion to get the real ip addres in case we put the host and not the ip directly
	//fmt.Println("ipAddrss VAlue: ", ipAddrss, "Error VAlue:", err)

	if err != nil {
		fmt.Println("Invalid Host, please try again...")
		panic(err)

		return nil, 0, "", err
	}

	//icmp.Message its an struct
	/*
		type Message struct {
			Type     Type        // type, either ipv4.ICMPType or ipv6.ICMPType
			Code     int         // code
			Checksum int         // checksum
			Body     MessageBody // body
	}*/

	// The type ipv4.ICMPTypeEcho is t send a request , Code: Not used for Echo and Echo Reply messages; set to 0,
	//For echo request de body is made of a Identifiers, Sequence NUmber and Payload
	//FOr the identifier is common to use the Id of the current proccess and use logic gate with en hex value
	//IN that way everytime that we restart the porcess itll be a differetn identifier

	icpmToSend := icmp.Message{Type: ipv6.ICMPTypeEchoRequest, Code: 0, Body: &icmp.Echo{ID: os.Getpid() & 0xffff, Seq: 1, Data: []byte(mssm)}}
	//fmt.Println("first id value", os.Getpid()&0xffff)

	//Marshal return the binary encoding of the Message
	binaryEncoding, err := icpmToSend.Marshal(nil)
	//fmt.Println("binaryEncoding value: ", (binaryEncoding), " Error value: ", err)

	if err != nil {

		return ipAddrss, 0, "", err
	}

	//Here well send the request
	//start := time.Now()

	/*
		WriteTo writes the ICMP message b to dst. The provided dst must be net.UDPAddr when c is a non-privileged datagram-oriented ICMP endpoint. Otherwise it must be net.IPAddr.
		from : https://godoc.org/golang.org/x/net/icmp#PacketConn
	*/

	start := time.Now()
	n, err := listener.WriteTo(binaryEncoding, ipAddrss)
	//fmt.Println("N value", n, " Error value: ", err)

	if err != nil {
		return ipAddrss, 0, "", err
	} else if n != len(binaryEncoding) {
		return ipAddrss, 0, "", fmt.Errorf("got %v; want %v", n, len(binaryEncoding))
	}

	// Wait for a reply
	reply := make([]byte, 1500)

	//Set the maximum time to wait for the response, in this case well wait 10 seconds
	err = listener.SetReadDeadline(time.Now().Add(time.Duration(timer) * time.Second))
	//fmt.Println(" Error value: ", err)

	if err != nil {

		return ipAddrss, 0, "", err
	}

	n, peer, err := listener.ReadFrom(reply)

	//fmt.Println("n value", n, " Error value: ", err, "Peer Value", peer)
	if err != nil {
		return ipAddrss, 0, "", err
	}
	duration := time.Since(start)

	//fmt.Println("duration value", duration)

	rm, err := icmp.ParseMessage(58, reply[:n])
	v := string(reply[8:n])

	//fmt.Println("rm value", v, "Error value: ", err, "reply: ")

	switch rm.Type {
	case ipv6.ICMPTypeEchoReply:
		return ipAddrss, duration, v, nil
	default:
		return ipAddrss, 0, "", fmt.Errorf("got %+v from %v; want echo reply", rm, peer)
	}
}

func ping(addr1 string, value int, mssm string, timer float64) {

	SetupCloseHandler()
	packetLoss := false
	if value == 4 {
		for {
			send = send + 1
			dst, dur, mss, err := getResponse(addr1, mssm, timer)
			log.Printf("Ping to (%s) IP Address: %s  TTl = %s\n", addr1, dst, dur)
			if mssm != mss {
				packetLoss = true
			}
			fmt.Printf("Original Message: %s    Received Message: %s    Packet loss:  %v\n\n", mssm, mss, packetLoss)

			if err == nil {
				received = received + 1

			}
			time.Sleep(2 * time.Second)

		}
	} else {
		for {
			send = send + 1
			dst, dur, mss, err := getResponsev6(addr1, mssm, timer)
			log.Printf("Ping to (%s) IP Address: %s  TTl = %s\n", addr1, dst, dur)
			if mssm != mss {
				packetLoss = true
			}
			fmt.Printf("Original Message: %s    Received Message: %s    Packet loss:  %v\n\n", mssm, mss, packetLoss)

			if err == nil {
				received = received + 1

			}
			time.Sleep(2 * time.Second)
		}
	}

}

//FUnction to hanlde Ctrl + c
func SetupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		x := math.Abs(received - send)
		loss = (x / send) * float64(100)

		fmt.Printf("\nPackages sent : %.0f   Packages received  : %.0f     Total loss  : %.0f %\n", send, received, loss)
		os.Exit(0)
	}()
}

var (
	protocolv4 bool
	protocolv6 bool
)

func main() {

	//addr1 := "google.com"
	protocolv4 := flag.Bool("4", true, "use IPv4 protocol")

	protocolv6 := flag.Bool("6", false, "use IPv6 protocol")

	mssm := flag.String("mssm", "Hello Cloudflare", "Define your own message to send")

	timer := flag.Float64("TTL", 10, "Time for TTL respone in seconds")

	flag.Usage = func() {
		fmt.Println("Usage: [options] host [Host Name/ Address]")
		fmt.Println("Options: ")
		flag.PrintDefaults()
	}

	flag.Parse()
	addr := flag.Args()[0]

	if *protocolv6 {
		ping(addr, 6, *mssm, *timer)
	} else if *protocolv4 {

		ping(addr, 4, *mssm, *timer)
	}

}
