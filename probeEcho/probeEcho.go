package probeEcho

import (
	"PRR_Labo4/networking"
	"fmt"
	"log"
	"math"
	"strconv"
)

const NB_PROCESS_ID_DIGITS = 3

// My id
var me int

// Id of my parent node
var parent int

// Amount of neighbors for this node
var nbNeighbors int

// List of the neighbors of this node
var neighbors []int

// Maps a calculation's unique id with the amount of probe/echos yet te be received for this message
var expectedMessages map[int]int

// Maps a calculation's unique id with the current possibility of it being prime. Built using the received echos
var canStillBePrime map[int]bool

// Maps the calculations that I started myself with the candidate of this calculation
var myRunningCalculations map[int]int

var localCalculationCurrentId = 0

var primeDivisor int

// Display some logs
var trace = false

type ProbeMessage struct {
	CalculationId int //the last 3 digits of CalculationId are the id of the original requester
	Parent        int
	Candidate     int
}

type EchoMessage struct {
	CalculationId int //the last 3 digits of CalculationId are the id of the original requester
	MayBePrime    bool
}

//The final result of a calculation
type Result struct {
	Candidate int //the last 3 digits of CalculationId are the id of the original requester
	IsPrime   bool
}

// Channel to request a new calculation
var InitNewCalculation = make(chan int)

var CalculationResult = make(chan Result)

// Channel to receive a new Probe message
var Probe = make(chan ProbeMessage)

// Channel to receive a new Echo message
var Echo = make(chan EchoMessage)

//// Channel to forward messages from the network
//var Message = make(chan CalculationMessage)

func Trace() {
	trace = true
}

func Run(processId int, nNeighbors int, neighborsArray []int, divisor int) {
	me = processId
	nbNeighbors = nNeighbors
	neighbors = neighborsArray
	primeDivisor = divisor

	for {
		select {
		case n := <-InitNewCalculation:
			newCalculation(n)
		case m := <-Probe:
			handleProbeMessage(m)
		case m := <-Echo:
			handleEchoMessage(m)
		}
	}
}

func handleProbeMessage(message ProbeMessage) {
	_, found := expectedMessages[message.CalculationId]
	if found {
		if trace {
			fmt.Println("[ProbeEcho] Received probe for an already existing calculation")
		}

		handleEchoMessage(EchoMessage{
			CalculationId: message.CalculationId,
			MayBePrime:    true,
		})

	} else {
		if trace {
			fmt.Println("[ProbeEcho] Received probe for a new calculation")
		}

		mayBePrime := (message.Candidate % primeDivisor) == 0

		if nbNeighbors > 1 {
			expectedMessages[message.CalculationId] = nbNeighbors - 1
			canStillBePrime[message.CalculationId] = mayBePrime

			probeMessage := networking.CalculationMessage{
				IsProbe:       true,
				CalculationId: message.CalculationId,
				Emitter:       me,
				Candidate:     message.Candidate,
			}

			//Forwarding the probes to the children
			for _, n := range neighbors {
				if n != parent {
					networking.SendMesage(n, probeMessage)
				}
			}

		} else { //If this node is a leaf, it directly sends an echo
			echoMessage := networking.CalculationMessage{
				IsProbe:       false,
				CalculationId: message.CalculationId,
				MayBePrime:    mayBePrime,
			}
			networking.SendMesage(parent, echoMessage)
		}

	}
}

func handleEchoMessage(message EchoMessage) {
	//update the amount of expected messages and the possibility of the candidate being prime
	expectedMessages[message.CalculationId]--
	canStillBePrime[message.CalculationId] = canStillBePrime[message.CalculationId] && message.MayBePrime

	//If we don't expect any more messages for this calculation, we send an echo with our results
	if expectedMessages[message.CalculationId] == 0 {

		messageOriginalProcessId := message.CalculationId % int(math.Pow10(NB_PROCESS_ID_DIGITS))

		//If I am the original sender of the calculation request, return the result
		if messageOriginalProcessId == me {
			CalculationResult <- Result{
				Candidate: myRunningCalculations[message.CalculationId],
				IsPrime:   false,
			}
			//If I am not the original sender of the calculation request, send an echo to the parent
		} else {
			echoMessage := networking.CalculationMessage{
				IsProbe:       false,
				CalculationId: message.CalculationId,
				MayBePrime:    canStillBePrime[message.CalculationId],
			}
			networking.SendMesage(parent, echoMessage)
		}

		//clean up the map entries
		delete(expectedMessages, message.CalculationId)
		delete(canStillBePrime, message.CalculationId)

	}
}

func newCalculation(candidate int) {

	//Checks that the process Id does not violate the IDs convention
	maxProcessId := int(math.Pow10(NB_PROCESS_ID_DIGITS))
	if me > maxProcessId-1 {
		log.Fatal("Error: To emit a new calculation, process Id must be lower than " + strconv.Itoa(maxProcessId))
		return
	}
	//Create the unique id for this calculation
	id := localCalculationCurrentId*maxProcessId + me
	localCalculationCurrentId++

	//If we can already find out that the candidate is not prime with our own divisor, directly return the result
	if candidate%primeDivisor == 0 {
		CalculationResult <- Result{
			Candidate: candidate,
			IsPrime:   false,
		}
		return
	}

	myRunningCalculations[id] = candidate
	expectedMessages[id] = nbNeighbors
	canStillBePrime[id] = true

	//Send the probe to all the neighbors
	probeMessage := networking.CalculationMessage{
		IsProbe:       true,
		CalculationId: id,
		Emitter:       me,
		Candidate:     candidate,
	}
	for _, n := range neighbors {
		networking.SendMesage(n, probeMessage)
	}
}
