package main

import "log"
import "os"
import "strconv"
import node "./utils/node_defs"
import msg "./utils/message_defs"
import join "./join_ring"
import leave "./leave_ring"
import init_ring_fingers "./init_ring_fingers"
import "math/rand"
import "sync"
import "time"
import "io/ioutil"
import "strings"
import "encoding/json"
import responsetime "./utils/responsetime"

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
var ring_nodes = make(map[int64](*node.Node))

/*
This is the global sync group to handle the goroutines properly
*/
var wg sync.WaitGroup


/*
This is the number of nodes in the ring
*/
var number_of_network_nodes int = 0

/*
This is the mean time for each node to wait before accepting the next message in the channel
*/
var mean_wait_value float64

func check_error(err error){
	if err != nil {
		log.Println("Error : ", err)
		os.Exit(1)
	}
	
}
/*Gets a random node in the chord ring
*/
func get_random_ring_node() (rand_num int64) {
    for(true){
        rand_num := rand.Intn(number_of_network_nodes)
        //If we generated a channel id that is in use in the ring, return the number
        if len(ring_nodes) == 0 {
            return -1
        }
        if val, ok := ring_nodes[int64(rand_num)]; ok {
            _ = val
            log.Printf("Random ring node:%d\n", int64(rand_num))
            return int64(rand_num)
        }
    }
    return int64(rand_num)
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
		//add node to network
		network[int64(id)] = make(chan string, 100)	
		//start up node
		id_64 := int64(id)
		wg.Add(1)
		go net_node(id_64)
	}
}



/*
This is a routine that defines a node. The routine listens on the channel that is assigned
to the given channel id  for incoming messages.
*/
func net_node(channel_id int64){
	
        defer wg.Done()
	//create a node structure to store information,
	//successor/predecessor references, etc.
	var node_obj = node.Node {ChannelId: channel_id, Successor:nil, Predecessor:nil}
	var wait_time = int(responsetime.GetResponseTime(mean_wait_value))

	//If ring is empty just add this node to the ring
	//This is the first node to enter the ring. Make this node's successor itself.
	//create
	if len(ring_nodes) == 0{
	node_obj.Successor = &node_obj
	ring_nodes[channel_id] = &node_obj
	log.Printf("Node %d was used to create the ring.", channel_id)
	}
	for {
		select {
			case <-time.After(time.Duration(wait_time) * time.Second):
				msg_recv := <-network[channel_id]
				//wait an average of AVERAGE_WAIT_TIME seconds before accepting a message
				log.Printf("\nWaiting %d seconds before processing message for Node: %d\n", wait_time, channel_id)
				wait_time = int(responsetime.GetResponseTime(mean_wait_value))
				time.Sleep(time.Duration(wait_time) * time.Second)
				log.Printf("\nNode: %d recieved the following message:%s\n", channel_id, msg_recv)

				byte_msg := []byte(msg_recv)
				var message msg.Message
				err := json.Unmarshal(byte_msg, &message)
				if err != nil {
					log.Printf("Node: %d failed to unmarshal the json string", channel_id)
					break
				}

				//Perform join-ring action
				if message.Do == "join-ring" {
					if val, ok := ring_nodes[channel_id]; ok != true {
						_ = val
						sponsoring_node_id := message.SponsoringNode
						join.Join_ring(sponsoring_node_id, &node_obj)
						ring_nodes[channel_id] = &node_obj
					}else{
						log.Printf("\nNode %d is already in the ring; cannot join-ring\n", channel_id)
					}
			   	} else if message.Do == "leave-ring" {
					if val, ok := ring_nodes[channel_id]; ok{
						_ = val
						leave.Leave_ring(&node_obj, message.Mode)
						delete(ring_nodes, channel_id)
					}else{
						log.Printf("\nNode %d is not in the ring; cannot leave-ring\n", channel_id)
					}
				}else if message.Do == "init-ring-fingers" {
					if val, ok := ring_nodes[channel_id]; ok{
						_ = val
						init_ring_fingers.Init_Ring_Fingers(&node_obj)
					}else{
						log.Printf("\nNode %d is not in the ring; cannot init-ring-fingers\n", channel_id)
					}
				}

				/*else if (message.Do == "put"){
					respond_to_node_id = struct_message.RespondTo
					data  = struct_message.Data
					put(data, respond_to_node_id, node_obj)
				}...
				*/
				//print_ring_nodes()
			default:
				time.Sleep(1)
				continue
		}
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
		log.Printf("\nNode %d is in the ring\n", channel_id)
	}
}

/*
This function reads in lines from a file that contain json
*/
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
	mean_wait, err := strconv.ParseFloat(prog_args[2], 64)
	check_error(err)
	mean_wait_value = mean_wait
	number_of_network_nodes = num_nodes
	log.Println("This is the coordinator.")

	//Create a bunch of random nodes for the network
	init_topology()

	//Get a random network nodes id
	var random_node_id int64

	//get a list of string json instructions to send to random nodes
	var instructions []string = create_message_list(file_name)
	for i := 0; i < len(instructions); i++ {
		//pick a random node in the ring to send the message to.
		random_node_id = get_random_ring_node()
		random_network_id := get_random_network_node()
			byte_msg := []byte(instructions[i])
			var message msg.Message
			err := json.Unmarshal(byte_msg, &message)
			if err != nil {
				log.Println("Reached the end of the json instructions")
				break
			}
			//format join ring instruction with random sponsoring node
			if message.Do == "join-ring" {

				if random_node_id > 0 {
					message.SponsoringNode = random_node_id
				}else{
					log.Println("There is no node to sponsor for join ring: %d")
					continue
				}
			}

			modified_inst, err := json.Marshal(message)
			log.Printf("Read the following instruction from file %s.", string(modified_inst))
			check_error(err)
			// Give a random node instructions 
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
		if len(prog_args) < 3 {
		log.Println("USAGE: go run main.go <INSTRUCTION FILE> <NUM NODES> <AVERAGE_WEIGHT_TIME>")
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


