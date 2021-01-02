package probeEcho

// My id
var me int

// Id of my parent node
var parent int

// Amount of neighbors for this node
var nbNeighbors int

// List of the neighbors of this node
var neighbors []int

// Maps a message's unique id with the amount of probe/echos yet te be received for this message
var ids map[int]int

// Display some logs
var trace = false

//type CalculationMessage struct {
//	IsProbe  		bool	//true = probe, false = echo
//	CalculationId 	int
//	Emitter			int 	//zero if the message is not a probe
//	Candidate 		string	//Empty if the message is not a probe
//}

type ProbeMessage struct {
	CalculationId int
	Parent        int
	Candidate     int
}

type EchoMessage struct {
	CalculationId int
	MayBePrime    bool
}

// Channel to request a new calculation
var InitNewCalculation = make(chan int)

// Channel to receive a new Probe message
var Probe = make(chan ProbeMessage)

// Channel to receive a new Echo message
var Echo = make(chan EchoMessage)

//// Channel to forward messages from the network
//var Message = make(chan CalculationMessage)

func Trace() {
	trace = true
}

func Run(processId int, nNeighbors int, neighborsArray []int) {
	me = processId
	nbNeighbors = nNeighbors
	neighbors = neighborsArray

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

}

func handleEchoMessage(message EchoMessage) {

}

func newCalculation(candidate int) {

}
