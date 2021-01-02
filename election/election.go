package election

import (
	"PRR_labo3/networking"
	"fmt"
	"strconv"
	"time"
)

/*
	Implementation of the Bully Algorithm
	The RunBullyAlgorithm must be run ONLY ONCE at application startup
	The other processes can then interact with the election algorithm by using
	the thread-safe channels
*/

// Number of processes in the topology
var nProcesses int

// My id
var me int

// My current aptitude
var aptitude int

// Other processes aptitudes
var apts []int

// Election in progress or not
var inProgress bool

// Chosen process after an election
var chosen int

// Number of requests retrieving the chosen blocked by an in progress election
var chosenRequestsWaiting int

// Max transmission time from one side of the topology to the other
var maxTransmissionDuration time.Duration

// Aptitude changed when an election is in progress
var aptitudeChanged bool

// Display some logs
var trace bool

type BullyMessage struct {
	Emitter  int
	Aptitude int
}

// Channel to request an election
var Election = make(chan struct{})

// Channel to request a read on the chosen
var RequestChosen = make(chan struct{})

// Channel to retrieve the chosen (must ONLY be called after a request)
var GetChosen = make(chan int)

// Channel to forward messages from the network
var Message = make(chan BullyMessage)

// Channel to set a new aptitude
var SetAptitude = make(chan int)

// Internal channel to handle the timeouts
var timeout = make(chan struct{})

func Trace() {
	trace = true
}

func RunBullyAlgorithm(processId int, totalProcesses int, currentApt int, maxTD time.Duration) {
	maxTransmissionDuration = maxTD
	me = processId
	nProcesses = totalProcesses
	aptitude = currentApt

	if trace {
		fmt.Println("[Bully] Waiting until all elections are done...")
	}
	inProgress = true
	for inProgress {
		timeout := time.After(maxTransmissionDuration)
		select {
		case <-Message:
		case <-timeout:
			inProgress = false
		}
	}

	apts = make([]int, nProcesses)
	chosen = me

	if trace {
		fmt.Println("[Bully] Enter primary loop")
	}
	for {
		select {
		case <-Election:
			if !inProgress {
				startElection()
			}
		case <-RequestChosen:
			handleChosenRequest()
		case m := <-Message:
			handleMessage(m)
		case <-timeout:
			handleTimeout()
		case apt := <-SetAptitude:
			handleSetAptitude(apt)
		}
	}
}

func startElection() {
	if trace {
		fmt.Println("[Bully] Election started")
	}
	inProgress = true
	for i := range apts {
		apts[i] = 0
	}
	apts[me] = aptitude

	networking.SendStartElection(me, aptitude)
	go waitForTimeout()
}

func handleChosenRequest() {
	if !inProgress {
		GetChosen <- chosen
	} else {
		chosenRequestsWaiting++
	}
}

func handleMessage(message BullyMessage) {
	if trace {
		fmt.Println("[Bully] Message received from " + strconv.Itoa(message.Emitter) + " with aptitude " + strconv.Itoa(message.Aptitude))
	}
	if !inProgress {
		if trace {
			fmt.Println("[Bully] No election in progress, starting election")
		}
		startElection()
	}
	apts[message.Emitter] = message.Aptitude
}

func handleTimeout() {
	if trace {
		fmt.Println("[Bully] A timeout occurred, choosing winner...")
	}
	for process, apt := range apts {
		if apt > apts[chosen] || (apt == apts[chosen] && process < chosen) {
			chosen = process
		}
	}
	inProgress = false

	// Free the processes waiting for an election result
	for i := 0; i < chosenRequestsWaiting; i++ {
		GetChosen <- chosen
	}
	chosenRequestsWaiting = 0

	// A process changed the aptitude during the election
	if aptitudeChanged {
		Election <- struct{}{}
		aptitudeChanged = false
	}
	if trace {
		fmt.Println("[Bully] Winner chosen: " + strconv.Itoa(chosen))
	}
}

func handleSetAptitude(apt int) {
	if trace {
		fmt.Println("[Bully] Aptitude changed to " + strconv.Itoa(apt))
	}
	aptitude = apt
	if inProgress {
		aptitudeChanged = true
	} else {
		go func() { Election <- struct{}{} }()
	}
}

func waitForTimeout() {
	timer := time.NewTimer(2 * maxTransmissionDuration)
	<-timer.C
	timeout <- struct{}{}
}
