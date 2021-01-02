package main

import (
	"PRR_labo3/election"
	"PRR_labo3/networking"
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var pingChannels []chan struct{}
var MaxTransmissionDuration time.Duration
var me int
var debug = false
var trace = false

func main() {

	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide Id")
		return
	}

	//Read the config file
	var configFile = "config.json"
	var topology, err = Parse(configFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	//Get the properties from the config file
	myId, _ := strconv.Atoi(arguments[1])
	myPort := topology.Clients[myId].Port
	myAddress := topology.Clients[myId].Hostname
	nbProcesses := topology.ClientCount
	startApt := topology.Clients[myId].Apt

	MaxTransmissionDuration = time.Millisecond * time.Duration(topology.MaxTD)
	me = myId

	//if Id is too high for the amount of processes, exit
	if myId >= nbProcesses {
		return
	}

	//Set the options if they are in the config.json file
	if topology.Debug {
		networking.Debug()
		debug = true
		MaxTransmissionDuration += 3 * time.Second
	}
	if topology.Trace {
		networking.Trace()
		election.Trace()
		trace = true
	}

	//Init addresses of all the processes
	addresses := make([]string, nbProcesses)
	for i, cli := range topology.Clients {
		addresses[i] = cli.Hostname + ":" + strconv.Itoa(cli.Port)
	}
	networking.SetAddresses(addresses)

	//init ping handling channels
	pingChannels = make([]chan struct{}, nbProcesses)
	for i := 0; i < nbProcesses; i++ {
		pingChannels[i] = make(chan struct{})
	}

	//Start the network-related goroutines
	go networking.Send()
	go networking.ListenMulticast(topology.MultiCastAddr, consumeMulticast)
	go networking.ListenUnicast(strconv.Itoa(myPort), myAddress, consumeUnicast)

	//start the probeEcho algorithm
	go election.RunBullyAlgorithm(myId, nbProcesses, startApt, MaxTransmissionDuration)
	election.Election <- struct{}{}

	//ping the current elect process periodically
	go periodicPings(myId)

	// Prompt and wait for manual aptitude updates
	mainloop()

}

// Wait for commands from the user
func mainloop() {
	promptForNewAptitude()
	var command string

	for {
		command = ""
		reader := bufio.NewReader(os.Stdin)

		//parse the command
		command, _ = reader.ReadString('\n')
		msg := strings.TrimSpace(command)
		newApt, err := strconv.Atoi(msg)
		if err != nil {
			fmt.Println("Please input a number")
		} else {
			election.SetAptitude <- newApt
			fmt.Println("Aptitude set to " + msg)
			promptForNewAptitude()
		}

	}
}

func promptForNewAptitude() {
	fmt.Println("Type a new aptitude to change it : ")
}

//Consume a message read from the unicast
func consumeUnicast(reader *bytes.Reader) {
	//msgType: 1 = PING, 2 = PONG
	msgType, _, err := reader.ReadRune()
	if err != nil {
		log.Fatal(err)
	}

	processId, _, err := reader.ReadRune()
	if err != nil {
		log.Fatal(err)
	}

	//interpret the message
	switch msgType {
	case 1: //ping
		fmt.Println("[Network] received ping")
		networking.SendPong(int(processId), me)
	case 2: //pong
		fmt.Println("[Network] received pong")
		pingChannels[processId] <- struct{}{}
	default:
		log.Fatal("Unknown message type")
	}
}

//Consume a message read from the multicast
func consumeMulticast(reader *bytes.Reader) {
	processId, _, err := reader.ReadRune()
	if err != nil {
		log.Fatal(err)
	}

	//Ignore if the message comes from myself
	if int(processId) == me {
		return
	}

	aptitude, _, err := reader.ReadRune()
	if err != nil {
		log.Fatal(err)
	}

	if trace {
		fmt.Println("[Network] Received start-probeEcho message")
	}

	//Request to start an probeEcho
	msg := election.BullyMessage{
		Emitter:  int(processId),
		Aptitude: int(aptitude),
	}
	election.Message <- msg
}

//Wait for a pong message after sending a ping
func waitForPong(processId int) {
	if trace {
		fmt.Println("[Network] Waiting for pong...")
	}
	timeout := time.After(2 * MaxTransmissionDuration)
	select {
	case <-pingChannels[processId]: //received a pong in time

	case <-timeout: //timed out
		if trace {
			fmt.Println("[Network] Pong waiting timed out, starting probeEcho...")
		}
		election.Election <- struct{}{}
	}
}

//Send a ping to the probeEcho winner periodically
func periodicPings(processId int) {
	for {
		time.Sleep(3 * time.Second)

		if debug {
			time.Sleep(10 * time.Second)
		}

		//Get the current elected process
		election.RequestChosen <- struct{}{}
		chosen := <-election.GetChosen
		if processId == chosen {
			continue
		}

		networking.SendPing(chosen, processId)
		go waitForPong(chosen)
	}

}
