# bully
An RPC-based implementation of the Bully algorithm for leader election in Golang.

**Description:**

This code implements the Bully algorithm for leader election in a distributed system using Go. The Bully algorithm is a method for electing a coordinator (leader) among a group of nodes (servers) in a distributed system. The leader is responsible for coordinating tasks and ensuring the system's overall health. In the given code, an RPC-based communication approach is used to connect and communicate between nodes.

Here's a breakdown of the code:

Imports: Import necessary Go packages, such as net/rpc, net/http, sync, time, os, bufio, fmt, log, and strconv.

Structs and variables: Define structs BullyNode, VoteArguments, VoteReply, HeartbeatArgument, HeartbeatReply, and ServerConnection. Also, define global variables selfID, serverNodes, electionTimer, electionTimeout, isLeader, isCand, and timeout.

RPC methods: Implement RPC methods ElectionMessage and Heartbeat for the BullyNode struct.

LeaderElection: Define the LeaderElection function, which initiates the election process by sending election messages to all other nodes.

Coordinator: Define the Coordinator function, which sends heartbeat messages to all other nodes when the current node is the leader.

main function: Initialize the server, read command-line arguments, read the cluster configuration file, register RPCs, establish connections with other servers, and start leader election and heartbeat threads.

Here is how our algorithm works:

1. Start separate goroutines for handling election timer expiration, leader election, leader state transition, and sending heartbeat messages when the current node is the leader.
2. Assign a random time out for failure detection (we did this so to mimic a realistic distributed system running the bully algorithm, because we are running our code locally, any failure would almost be simulationously observed by all nodes)
3. Once failure is detected, the node that detected failure will set itself as candidate and start the election
4. The candidate will send election messages (RPCs) to all nodes
5. nodes with higher value will get their failure timers reset and set themselves as candidate, they will also set the RPC value isGreater to true (nodes with lower value with just get their failure timers reset)
6. if the canidate from 4 gets a true isGreater, they set themselves as no longer a candidate
7. If there are new candidates, the new candidates from 5 will start election again from step 4.
8. If there are no new candidates/no higher alive nodes aka the constant timeout is up and the node is the candidate, that node becomes leader.
9. heartbeat messages start, these only reset the failure timers because we do not have content to send.

**How To Run:** 
1. Download both the bully.py file and the cluster.txt file. Similarly to our Raft implementation, the cluster.txt file will tell the code which nodes should be avalible to connect to. This cluster.txt file is set up to be run locally all on the same device, however, it could be simply editted to work on an distributed network without edits to bully.py.
2. When ready to run the code, go into the command line of the device and open 5 tabs.
3. For each node/tab in the command line, enter the command: go run /path/to/bully.go node# /path/to/cluster.txt
4. Your node # should be a unique number from 0 to 4 inclusive.
5. You should see a prompt that tells you to "Type anything when ready to connect >>"
6. once every tab/node has had that command entered, type anything into each command to get the bully algorithm started.
7. Congrats! You are now running the bully algorithm :)
