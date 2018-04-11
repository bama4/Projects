package main

import "fmt"
import "os"
import "strconv"
import node "./utils/node_defs"
import msg "./utils/message_defs"
import "math/rand"
import "sync"
import "time"
import "io/ioutil"
import "strings"
import "encoding/json"

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
var ring_nodes = make(map[int64](node.Node))

/*
This is the global sync group to handle the goroutines properly
*/
var wg sync.WaitGroup


/*
This is the number of nodes in the ring
*/
var number_of_network_nodes int = 0

func check_error(err error){
	if err != nil {
		fmt.Println("Error : ", err)
		os.Exit(1)
	}
	
}
/*Gets a random node in the chord ring
*/
func get_random_ring_node() (rand_num int64) {
	for(true){
		rand_num := rand.Intn(number_of_network_nodes)
		//If we generated a channel id that is in use in the ring, return the number
		fmt.Printf("%d\n", rand_num)
		if len(ring_nodes) == 0 {
			fmt.Println("The chord ring is empty")
			return -1
		}
		if val, ok := ring_nodes[int64(rand_num)]; ok {
			_ = val
			fmt.Printf("%d\n", rand_num)
			return int64(rand_num)
		}
	}
	return int64(rand_num)
}

/*Adds a random node to the ring if the ring is empty
*/
func create_ring(){

	if len(ring_nodes) == 0 {
		rand_net_id := get_random_network_node()
		network[rand_net_id] <- "{'do':'create'}"
	}
}

/*Gets a random node in the network
*/
func get_random_network_node() (rand_num int64){
	for(true){
		rand_num := rand.Intn(number_of_network_nodes)
		//If we generated a channel id that is in use in the network, return the number
		if val, ok := network[int64(rand_num)]; ok {
			_ = val
			return int64(rand_num)
		}
	}
	return int64(rand_num)
}

/*
Generates a unique channel id that is not already in the network
*/
func generate_channel_id() (rand_num int64){

	rand_num = 0
	if len(network) == number_of_network_nodes {
		//cant generate a unique id
		return -1
	}

	for(true){
		rand_num := rand.Intn(number_of_network_nodes)
		//If we generated a channel id that is not in use
		//, return True
		if val, ok := network[int64(rand_num)]; ok != true {
			_ = val
			return int64(rand_num)
		}
	}
	return int64(rand_num)
}


/*
Initializes the network with nodes with random identifiers.
Creates nodes with random identifiers and adds them to the network map.
*/
func init_topology(){
	

	for i:=0; i < number_of_network_nodes; i++ {
		id := generate_channel_id()
		
		fmt.Printf("%d\n", id)
	
		//add node to network
		network[int64(id)] = make(chan string, 100)	
		//start up node
		id_64 := int64(id)
		wg.Add(1)
		go net_node(id_64)
	}

		//randomly add a node to the chord network
		create_ring()
}


/*
This is a routine that defines a node. The routine listens on the channel that is assigned
to the given channel id  for incoming messages.
*/
func net_node(channel_id int64){
	
        defer wg.Done()
	//create a node structure to store information,
	//successor/predecessor references, etc.
	var node_obj = node.Node {ChannelId: channel_id}
	var is_in_ring = false

	for {
		select {
			case msg_recv := <-network[channel_id]:

				fmt.Printf("\nNode: %d\n", channel_id)
				fmt.Println("Message Recieved: ", msg_recv)
				if msg_recv == "{'do':'create'}" {
					//This is the first node to enter the ring. Make this node's successor itself.
					node_obj.Successor = &node_obj
					ring_nodes[channel_id] = node_obj
					is_in_ring = true
					fmt.Printf("Node %d is in the ring now. %b", channel_id, is_in_ring)
				}
				/*
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
				*/
				return
			default:
				time.Sleep(3)
				return
		}
		return
	}
}

/*
/*A function that cleans up after goroutines*/
func cleanup(){
	for _, channel := range network {
		close(channel)
	}
}

func print_ring_nodes(){
	for channel_id, _ := range ring_nodes {
		fmt.Print("Channel id %s is in the ring", channel_id)
	}
}


func create_message_list(file_name string) []string {
	dat, err := ioutil.ReadFile(file_name)
	check_error(err)
	data := string(dat)
	var inst_list []string = strings.Split(data, "\n")
	return inst_list
}

func coordinator(prog_args []string){

	var file_name = prog_args[0]
	_ = file_name
	var num_nodes, _ = strconv.Atoi(prog_args[1])
	number_of_network_nodes = num_nodes
	fmt.Println("This is the coordinator.")

	//Create a bunch of random nodes for the network
	init_topology()

	//Send a message
	var random_node_id = get_random_network_node()
	network[random_node_id] <- "{'do':'something'}"
	//get a list of string json instructions to send to random nodes
	var instructions []string = create_message_list(file_name)
	for i := 0; i < len(instructions); i++ {
		//pick_random_net_node() pick a random node on network to send the message to.
		random_node_id = get_random_ring_node()
		random_network_id := get_random_network_node()
		fmt.Printf("Read the following instruction from file %s", instructions[i])
			byte_msg := []byte(instructions[i])
			var message msg.Message
			err := json.Unmarshal(byte_msg, &message)
			if err != nil {
				fmt.Println("Reached the end of the json instructions")
				break
			}
			//format join ring instruction with random sponsoring node
			if message.Do == "join-ring" {

				if random_node_id > 0 {
					message.SponsoringNode = strconv.FormatInt(random_node_id, 10)
				}else{
					fmt.Println("There is no node to sponsor for join ring")
					continue
				}
			}

			modified_inst, err := json.Marshal(message)
			check_error(err)
			fmt.Printf("Instruction in file updated to: %s", string(modified_inst))
			// Tell a random node to join the chord ring
			network[random_network_id] <- string(modified_inst) 
	}
	
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
		fmt.Println("USAGE: go run main.go <INSTRUCTION FILE> <NUM NODES>")
		os.Exit(1)
	}

	//Set up random generator seed
	rand.Seed(time.Now().UTC().UnixNano())

	//run coordinator
	coordinator(prog_args)
	wg.Wait()
	cleanup()
	return
}

