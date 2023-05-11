# bully
An RPC-based implementation of the Bully algorithm for leader election in Golang.

**How To Run:** 
1. Download both the bully.py file and the cluster.txt file. Similarly to our Raft implementation, the cluster.txt file will tell the code which nodes should be avalible to connect to. This cluster.txt file is set up to be run locally all on the same device, however, it could be simply editted to work on an distributed network without edits to bully.py.
2. When ready to run the code, go into the command line of the device and open 5 tabs.
3. For each node/tab in the command line, enter the command: go run /path/to/bully.go node# /path/to/cluster.txt
4. Your node # should be a unique number from 0 to 4 inclusive.
5. You should see a prompt that tells you to "Type anything when ready to connect >>"
6. once every tab/node has had that command entered, type anything into each command to get the bully algorithm started.
7. Congrats! You are now running the bully algorithm :)

**Description:**

This code implements the Bully algorithm for leader election in a distributed system using Go. The Bully algorithm is a method for electing a coordinator (leader) among a group of nodes (servers) in a distributed system. The leader is responsible for coordinating tasks and ensuring the system's overall health. In the given code, an RPC-based communication approach is used to connect and communicate between nodes.

Here's a breakdown of the code:

Imports: Import necessary Go packages, such as net/rpc, net/http, sync, time, os, bufio, fmt, log, and strconv.

Structs and variables: Define structs BullyNode, VoteArguments, VoteReply, HeartbeatArgument, HeartbeatReply, and ServerConnection. Also, define global variables selfID, serverNodes, electionTimer, electionTimeout, isLeader, isCand, and timeout.

RPC methods: Implement RPC methods ElectionMessage and Heartbeat for the BullyNode struct.

LeaderElection: Define the LeaderElection function, which initiates the election process by sending election messages to all other nodes.

Coordinator: Define the Coordinator function, which sends heartbeat messages to all other nodes when the current node is the leader.

main function: Initialize the server, read command-line arguments, read the cluster configuration file, register RPCs, establish connections with other servers, and start leader election and heartbeat threads.

**Synchronization**

The problem:

The Bully algorithm relies on the assumption that all nodes are synchronized, making it a slightly less practical algorithm to implement in real life than, for example, the Ring algorithm. The Ring algorithm is an asynchronous time algorithm, as nodes do not need to be on the same clock in order for the algorithm to detect failures or to otherwise function properly. The reason that the Bully algorithm requires synchronization is because a large part of the algorithm relies on global timeouts. For example, let us envision a network of 6 nodes. Node 6 is the leader, but it fails. The failure is detected by node 3 by using the notion of a timeout, which allows node 3 to assume that node 6 is dead given that it does not respond to routine heartbeat messages for longer than a pre decided-upon measure of time. Nodes being structured in this way can be achieved if the system is running on a synchronized clock, which is difficult to achieve in a truly distributed environment. 

Possible solutions we considered:

There are multiple ways to tackle the problem of synchronization in a synchronous network. As mentioned above, one of the main concerns with synchronization is making sure that all of the nodes are running on the same clock/time. To do this, we have to implement some sort of method of having the code run so that the different nodes can be on the same schedule or else there would be no synchronization. While there are multiple ways to implement such a clock in golang, we mainly looked at two options. The first was to use the system clock (the clock on the device running the code) to keep track of the time. This would allow us to make sure all of the nodes were looking at the same exact clock since every system has the same clock. This is ideal for a Bully Algorithm that is being run locally since all of the nodes would be on the same device and therefore have the same time. However, this approach becomes problematic when considering the potential for expanding the code beyond just local implementation and the possibility that different devices/servers would have different time metrics which would cause the different nodes to fall out of sync. The other approach was to use a package that would provide the same global time to all of the nodes. Luckily, GoLang has such a package which would allow the different nodes to get access to the same exact clock. This method allows for synchronization across devices because any server would be able to access the time from this package and the time would be universal for all devices.

We also considered possible values for the global timeout, trying to factor in the notion of message delay and predicting how long it would take for messages to reasonably reach each other. We did some research (presented earlier, in our project proposal writeup) in order to identify a timeout value we could use as a starting point. We found that According to the Loyola University Chicago, “The cross-continental US round trip delay is typically around 50-100 ms (propagation speed 200 km/ms in cable, 5,000-10,000 km cable route, or about 3-6000 miles)”, but delay is typically dependent on message size and transmission capacity. We also learned in class that worst-case delay with no failures is considered 5 * message transmission time, and since the worst case presented in the round trip delay research we did is 100ms, we decided to begin with a baseline timeout of 5 x 100=500ms. We knew this value would be subject to change as we were not aware of our system’s transmission capacity, so we used this number as the starting point and steadily increased until our timeout value yielded successful results.

What we did and our final approach:

We chose the GoLang time package instead of a global clock to ensure the maintenance of synchronization across, not only processes ran on the same local machine, but also on potential virtual machines. 

In the .go file, we initialized a global constant timeout variable that all processes will have access to. For every node declaring itself as a candidate, and every instance call to LeaderElection(), we kept track of the start time of that call. Then using time.now(), we compared the current time to the time at the initiation of leader election for each node and if timeout was ever reached, the node would be able to assume that it has become the leader as no higher nodes responded. If a higher node responded, the initial candidate node would stop its election process and would no longer check for timeouts. 

It is important to note that because we wanted to simulate a true distributed system, we did implement a random timeout for failure detection which is based on a localized timer for each node. This timer is reset every time there is communication with another node. In an actual Bully system, this value would not be randomized and the nodes would not behave in this manner. However, because we do not have a truly distributed system with actual messages being sent, this mimics how such a system would behave time wise. The reason why we do not have a system that realistically implements every element of the Bully algorithm is because our code is being run locally, so any failure would be essentially instantaneously recognized by all other nodes at the same time. 

You can see a more detailed technical breakdown of our approach in the description below.

**How Our Algorithm Works**

1. Start separate goroutines for handling election timer expiration, leader election, leader state transition, and sending heartbeat messages when the current node is the leader.
2. Assign a random time out for failure detection (we did this so to mimic a realistic distributed system running the bully algorithm, because we are running our code locally, any failure would almost be simultaneously observed by all nodes).
3. Once failure is detected, the node that detected failure will set itself as the candidate and start the election.
4. The candidate will send election messages (RPCs) to all nodes.
5. Nodes with higher value will get their failure timers reset and set themselves as candidate, they will also set the RPC value isGreater to true (nodes with lower value will just have their failure timers reset).
6. If the candidate from 4 gets a true isGreater, they set themselves back to follower status and are no longer a candidate.
7. If there are new candidates, the new candidates from 5 will start election again from step 4.
8. If there are no new candidates/no higher alive nodes aka the constant timeout is up and the node is the candidate, that node becomes leader.
9. Heartbeat messages start, these only reset the failure timers because we do not have content to send.
