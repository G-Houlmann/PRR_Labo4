package main

import (
	"PRR_Labo4/networking"
	"PRR_Labo4/probeEcho"
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

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
	primeDivisor := topology.Clients[myId].PrimeDivisor
	neighbors := topology.Clients[myId].Neighbors
	nbProcesses := topology.ClientCount

	me = myId

	//Set the options if they are in the config.json file
	if topology.Debug {
		networking.Debug()
		debug = true
	}
	if topology.Trace {
		networking.Trace()
		probeEcho.Trace()
		trace = true
	}

	//Init addresses of all the processes
	addresses := make([]string, nbProcesses)
	for i, cli := range topology.Clients {
		addresses[i] = cli.Hostname + ":" + strconv.Itoa(cli.Port)
	}
	networking.SetAddresses(addresses)

	if trace {
		fmt.Println("process ID = " + strconv.Itoa(me))
		fmt.Println("prime divisor = " + strconv.Itoa(primeDivisor))
	}

	//Start the network-related goroutines
	go networking.StartSending()
	go networking.ListenUnicast(strconv.Itoa(myPort), myAddress, consumeUnicast)

	//start listening for results
	go listenForResult()

	//start the probeEcho algorithm
	go probeEcho.Run(me, len(neighbors), neighbors, primeDivisor)

	// Prompt and wait for calculation requests
	mainloop()

}

// Wait for commands from the user
func mainloop() {
	//TODO gérer si on entre un nombre trop haut pour le nombre de processus
	promptForNewCalculationRequest()
	var command string

	for {
		command = ""
		reader := bufio.NewReader(os.Stdin)

		//parse the command
		command, _ = reader.ReadString('\n')
		msg := strings.TrimSpace(command)
		candidate, err := strconv.Atoi(msg)
		if err != nil {
			fmt.Println("Please input an integer")
		} else {
			probeEcho.InitNewCalculation <- candidate
			promptForNewCalculationRequest()
		}

	}
}

func listenForResult() {
	for {
		result := <-probeEcho.CalculationResult
		if result.IsPrime {
			fmt.Println(strconv.Itoa(result.Candidate) + " is a prime number!")
		} else {
			fmt.Println(strconv.Itoa(result.Candidate) + " is not a prime number...")
		}
	}
}

//Consume a message read from the unicast
func consumeUnicast(payload []byte) {

	message := networking.CalculationMessage{}

	err := json.Unmarshal(payload, &message)
	if err != nil {
		log.Fatal(err)
	}

	if message.IsProbe {
		if trace {
			fmt.Println("Received a probe message")
		}
		probeEcho.Probe <- probeEcho.ProbeMessage{
			CalculationId: message.CalculationId,
			Parent:        message.Emitter,
			Candidate:     message.Candidate,
		}
	} else {
		if trace {
			fmt.Println("Received an echo message")
		}
		probeEcho.Echo <- probeEcho.EchoMessage{
			CalculationId: message.CalculationId,
			MayBePrime:    message.MayBePrime,
		}
	}
}

func promptForNewCalculationRequest() {
	fmt.Println("Type an integer to find out whether it is prime : ")
}
