package networking

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

var addresses []string
var myConnection net.UDPAddr
var debug = false
var trace = false

type MessageToSend struct {
	Content  []byte
	DestAddr string
}

type CalculationMessage struct {
	IsProbe       bool `json:"isProbe"` //true = probe, false = echo
	CalculationId int  `json:"calculationId"`
	Emitter       int  `json:"emitter"`   //zero if the message is not a probe
	Candidate     int  `json:"candidate"` //Empty if the message is not a probe
	MayBePrime    bool `json:"mayBePrime"`
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
func ListenUnicast(localPort string, localAddress string, consume func(payload []byte)) {
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

		go consume(inputBytes[:length])

	}
}

//continuously send the messages received through the toSend channel
func StartSending() {
	conn, _ := net.ListenUDP("udp", &myConnection)
	defer conn.Close()
	for {
		msgToSend := <-toSend
		destAddress, err := net.ResolveUDPAddr("udp", msgToSend.DestAddr)
		if err != nil {
			log.Fatal(err)
		}

		if debug {
			time.Sleep(5 * time.Second)
		}
		_, err = conn.WriteTo(msgToSend.Content, destAddress)
		if err != nil {
			log.Fatal(err)
		}
	}
}

//Send a message through the network
func SendMessage(destProcessId int, message CalculationMessage) {
	destAddr := addresses[destProcessId]
	if trace {
		var messageType string
		if message.IsProbe {
			messageType = "probe"
		} else {
			messageType = "echo"
		}
		fmt.Println("[Network] Sending " + messageType + " message to " + destAddr)
	}

	p, err := json.Marshal(message)
	if err != nil {
		log.Fatal(err)
	}

	msg := MessageToSend{
		Content:  p,
		DestAddr: destAddr,
	}

	toSend <- msg
}
