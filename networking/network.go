package networking

import (
	"bytes"
	"fmt"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"time"
)

var addresses []string
var myConnection net.UDPAddr
var multicastAddr string
var debug = false
var trace = false

type MessageToSend struct {
	SenderProcessId int
	Content         []byte
	Receiver        int
	DestAddr        string
}

var toSend = make(chan MessageToSend)

func SetAddresses(newAddresses []string) {
	addresses = newAddresses
}

func Debug() {
	debug = true
}

func Trace() {
	trace = true
}

//Listen on UDP multicast
func ListenMulticast(Addr string, consume func(reader *bytes.Reader)) {
	multicastAddr = Addr

	addr, err := net.ResolveUDPAddr("udp", Addr)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenPacket("udp", Addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	p := ipv4.NewPacketConn(conn)
	err = p.JoinGroup(nil, addr)
	if err != nil {
		log.Fatal(err)
	}

	buffsize := 128
	buffer := make([]byte, buffsize)
	reader := bytes.NewReader(buffer)

	for {
		_, _, err := conn.ReadFrom(buffer)
		if err != nil {
			log.Fatal(err)
		}
		go consume(reader)

		reader.Reset(buffer)
	}
}

//Listen on UDP multicast
func ListenUnicast(localPort string, localAddress string, consume func(reader *bytes.Reader)) {
	conn, err := net.ListenPacket("udp", localAddress+":"+localPort)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	if err != nil {
		log.Fatal(err)
	}

	for {
		inputBytes := make([]byte, 256)
		length, _, _ := conn.ReadFrom(inputBytes)
		if err != nil {
			log.Fatal(err)
		}

		buffer := bytes.NewBuffer(inputBytes[:length])
		reader := bytes.NewReader(buffer.Bytes())
		go consume(reader)

		reader.Reset(inputBytes)
	}
}

//Send a message via udp
func Send() {
	conn, _ := net.ListenUDP("udp", &myConnection)
	defer conn.Close()
	for {
		msgToSend := <-toSend
		//if trace {fmt.Println("[Network] Sending udp to " + msgToSend.DestAddr)}
		destAddress, err := net.ResolveUDPAddr("udp", msgToSend.DestAddr)
		if err != nil {
			log.Fatal(err)
		}

		if debug {
			time.Sleep(3 * time.Second)
		}
		_, err = conn.WriteTo(msgToSend.Content, destAddress)
		if err != nil {
			log.Fatal(err)
		}
	}
}

//Send a pong message
func SendPong(destProcessId int, myProcessId int) {
	destAddr := addresses[destProcessId]
	if trace {
		fmt.Println("[Network] Sending pong to " + destAddr)
	}
	rs := []rune{2, rune(myProcessId)}

	msg := MessageToSend{
		SenderProcessId: myProcessId,
		Content:         []byte(string(rs)),
		Receiver:        destProcessId,
		DestAddr:        destAddr,
	}

	toSend <- msg
}

//Send a ping message
func SendPing(destProcessId int, myProcessId int) {
	destAddr := addresses[destProcessId]
	if trace {
		fmt.Println("[Network] Sending ping to " + destAddr)
	}
	rs := []rune{1, rune(myProcessId)}

	msg := MessageToSend{
		SenderProcessId: myProcessId,
		Content:         []byte(string(rs)),
		Receiver:        destProcessId,
		DestAddr:        destAddr,
	}

	toSend <- msg
}

//Send via multicast a start probeEcho message
func SendStartElection(processId int, aptitude int) {
	if trace {
		fmt.Println("[Network] Sending start-probeEcho message to " + multicastAddr)
	}
	rs := []rune{rune(processId), rune(aptitude)}

	msg := MessageToSend{
		SenderProcessId: processId,
		Content:         []byte(string(rs)),
		DestAddr:        multicastAddr,
	}

	toSend <- msg
}
