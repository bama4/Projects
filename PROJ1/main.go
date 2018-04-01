package main

import "fmt"
import "strconv"
import msg "./utils/message_defs"
import node "./utils/node_defs"
import "math/rand"

/*
This is the global "network" variable which is essentially a
map of all of the identifier/channel (aka network address/node) pairs
that we will have in our network.
The network contains the channel-id (node identifier which is randomly generated)
and the channel associated with that Id
*/
var network = make(map[int64](chan string))

/*
This is a map that consists of all of the nodes that are in the Chord ring
*/
var ring_nodes = make(map[int64](node.Node*))

/*
Generates a unique channel id that is not already in the network
*/
func generate_channel_id(max_id int) (rand_num int){

	/*
	var rand_num := 0
	if len(network) == max_id{
		//cant generate a unique id
		return nil
	}

	while(True){
		rand_num = rand.Intn(max_id)
		if rand_num not in network {
			return rand_num
		}
	}
	*/
}

/*
Initializes the network with nodes with random identifiers.
Creates nodes with random identifiers and adds them to the network map.
*/
func init_topology(num_nodes int){
	
	/*
	for i=0; i < num_nodes; i++ {
		id = generate_channel_id(1000)
		//add node to network
		network[id] = make(chan string)	
		//start up node
		go net_node(id)
	}
	*/
}



/*
This is a routine that defines a node. The routine listens on the channel that is assigned
to the given channel id  for incoming messages.
*/
/*
func net_node(channel_id){
	//create a node structure to store information,
	//successor/predecessor references, etc.
	var node_obj = node.Node {ChannelId: channel_id}

	for true {
		select {
			case msg_recv := <-network[channel_id]:
				//unmarshall string into struct object
				//based on message do a blocking action
				//struct_message = json.UnMarshall(msg_recv)
				//var action = struct_message.Do
				if action == "join"{
					sponsoring_node_id = struct_message.SponsoringNode
					join(sponsoring_node_id, node_obj)
				} else if (action == "put"){
					respond_to_node_id = struct_message.RespondTo
					data = struct_message.Data
					put(data, respond_to_node_id, node_obj)

				
				}...

	}
}
*/
	/*
func coordinator(prog_args []string){

	var file_name = prog_args[0]
	var num_nodes = int(prog_args[1])
	fmt.Println("This is the coordinator.")

	//Create a bunch of random nodes for the network
	init_topology(num_nodes)

	//get a list of string json instructions to send to random nodes
	var instructions := create_message_list(file_name)

	
	
	*/
}

/*
This is the main function which takes in the parameters for the program.
The parameters are the instruction file with each line containing
json-formatted instruction messages, the number of nodes to generate for the network, 
and (TODO) the mean variable to use in the randomization of the node response times.
*/
func main(){

	var prog_args = os.Args[1:]
		if len(prog_args) < 1 {
		log.Println("USAGE: go run main.go <INSTRUCTION FILE> <NUM NODES>")
		os.Exit(1)
	}

	//coordinator(prog_args)
}

