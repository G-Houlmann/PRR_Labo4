package networking

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

//addresses of the other processes. Need to be initialized with SetAddresses.
var addresses []string

//local UDP address
var myConnection net.UDPAddr

//Wait a few seconds between sending message
var debug = false

//Display debugging logs during the program's execution
var trace = false

//Struct to communicate a message to be sent through the network
type MessageToSend struct {
	Content  []byte
	DestAddr string
}

//Single struct transmitted through the network for this program.
//It has the fields for either a probe message or an echo message, and a boolean to identify it's type.
type CalculationMessage struct {
	IsProbe       bool `json:"isProbe"` //true = probe, false = echo
	CalculationId int  `json:"calculationId"`
	Emitter       int  `json:"emitter"`   //zero if the message is not a probe
	Candidate     int  `json:"candidate"` //Empty if the message is not a probe
	MayBePrime    bool `json:"mayBePrime"`
}

var toSend = make(chan MessageToSend)

//Set the addresses of the other processes
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
			time.Sleep(3 * time.Second)
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

	//Encode the payload to json
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
