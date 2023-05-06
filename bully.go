package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"time"
	"math/rand"
	"sync"
)

type BullyNode int

type VoteArguments struct {
	CandidateID int
}

type VoteReply struct {
	IsGreater bool
}

type HeartbeatArgument struct {
	LeaderID int
}

type HeartbeatReply struct {
	Success bool
}

type ServerConnection struct {
	serverID      int
	Address       string
	rpcConnection *rpc.Client
}

var selfID int
var serverNodes []ServerConnection
var electionTimer *time.Timer
var electionTimeout int
var isLeader bool
var isCand bool
var timeout time.Time


func (*BullyNode) ElectionMessage(arguments VoteArguments, reply *VoteReply) error {
	
	if (selfID>arguments.CandidateID){
		log.Println("set as cand")
		isCand = true
		reply.IsGreater = true
		timeout = time.Now().Add(time.Millisecond*5000)
	}
		
	electionTimer.Reset(time.Duration(electionTimeout) * time.Millisecond)

	return nil
}

func (*BullyNode) Heartbeat(arguments HeartbeatArgument, reply *HeartbeatReply) error {
	log.Println("heartbeat recieved")
	isCand = false
	isLeader = false

	electionTimer.Reset(time.Duration(electionTimeout) * time.Millisecond)//reset timer for followers

    reply.Success = true

    return nil
}

func LeaderElection() {
	input := VoteArguments {selfID}

	for _,server := range serverNodes{
		var reply VoteReply
		electionResults := server.rpcConnection.Go("BullyNode.ElectionMessage", input, &reply, nil)
		<-electionResults.Done
		//log.Println(electionResults.Reply.(*VoteReply).IsGreater)
		if electionResults.Reply.(*VoteReply).IsGreater == true{
			isCand=false
		}
		electionTimer.Reset(time.Duration(electionTimeout) * time.Millisecond)
	}
}


func Coordinator() {
	input := HeartbeatArgument {selfID}

	var reply HeartbeatReply

	for _,server := range serverNodes{
		server.rpcConnection.Go("BullyNode.Heartbeat", input, &reply, nil)
	}
}

func main() {
	// The assumption here is that the command line arguments will contain:
	// This server's ID (zero-based), location and name of the cluster configuration file
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide cluster information.")
		return
	}

	// Read the values sent in the command line

	// Get this sever's ID (same as its index for simplicity)
	myID, err := strconv.Atoi(arguments[1])
	// Get the information of the cluster configuration file containing information on other servers
	file, err := os.Open(arguments[2])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	myPort := "localhost"

	// Read the IP:port info from the cluster configuration file
	scanner := bufio.NewScanner(file)
	lines := make([]string, 0)
	index := 0
	for scanner.Scan() {
		// Get server IP:port
		text := scanner.Text()
		log.Printf(text, index)
		if index == myID {
			myPort = text
			index++
			continue
		}
		// Save that information as a string for now
		lines = append(lines, text)
		index++
	}
	// If anything wrong happens with readin the file, simply exit
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// Following lines are to register the RPCs of this object of type BullyNode
	api := new(BullyNode)
	err = rpc.Register(api)
	if err != nil {
		log.Fatal("error registering the RPCs", err)
	}
	rpc.HandleHTTP()
	go http.ListenAndServe(myPort, nil)
	log.Printf("serving rpc on port" + myPort)

	// This is a workaround to slow things down until all servers are up and running
	// Idea: wait for user input to indicate that all servers are ready for connections
	// Pros: Guaranteed that all other servers are already alive
	// Cons: Non-realistic work around

	 reader := bufio.NewReader(os.Stdin)
	 fmt.Print("Type anything when ready to connect >> ")
	 text, _ := reader.ReadString('\n')
	 fmt.Println(text)

	// Idea 2: keep trying to connect to other servers even if failure is encountered
	// For fault tolerance, each node will continuously try to connect to other nodes
	// This loop will stop when all servers are connected
	// Pro: Realistic setup
	// Con: If one server is not set up correctly, the rest of the system will halt

	for index, element := range lines {
		// Attempt to connect to the other server node
		client, err := rpc.DialHTTP("tcp", element)
		// If connection is not established
		for err != nil {
			// Record it in log
			log.Println("Trying again. Connection error: ", err)
			// Try again!
			client, err = rpc.DialHTTP("tcp", element)
		}
		// Once connection is finally established
		// Save that connection information in the servers list
		serverNodes = append(serverNodes, ServerConnection{index, element, client})
		// Record that in log
		fmt.Println("Connected to " + element)
	}

	// Start leader election and heartbeat threads
	heartbeatInterval := 200 // milliseconds
	electionTimeout = (rand.Intn(10)+10)*500 // milliseconds

	selfID = myID
	isCand = false
	isLeader = false

	//election timer
	electionTimer = time.NewTimer(time.Duration(electionTimeout) * time.Millisecond)
	timeout = time.Now().Add(time.Millisecond*5000)
	wg := sync.WaitGroup{}
	wg.Add(2)


	go func() {
		for{
		select {
			case <- electionTimer.C:
				// Failure Detection timeout, node is now candidate 
				log.Println("failure detected")
				//LeaderElection()
				timeout = time.Now().Add(time.Millisecond*5000)
				log.Println("set as cand")
				isCand = true
				// Reset the timer for the next election
				electionTimer.Reset(time.Duration(electionTimeout) * time.Millisecond)
		}
	}
	}()

	go func() {
		for{
			if isCand{
				//log.Println("election started")
				//timeout = time.Now().Add(time.Millisecond*5000)
				LeaderElection()
			}
		}
	}()

	
	go func() {
		for {
			if (timeout.Before(time.Now()) && isCand){
				log.Println("leader time")
				isCand = false
				isLeader = true
			}
		}
	}()

	go func() {
		for{
			if isLeader{
				time.Sleep(time.Duration(heartbeatInterval) * time.Millisecond)
				// Send heartbeat messages to all other servers
				log.Println("coord sent")
				Coordinator()
				electionTimer.Reset(time.Duration(electionTimeout) * time.Millisecond)
			}
		}
	}()

	wg.Wait()

}
