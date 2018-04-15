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

/*This is a lock that should eb used when writing to maps
*/
var map_lock = sync.Mutex{}


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

/*Maps a string to an identifier
*/
func map_string_to_id(msg string)(identifier int64){
	var msg_sum int64 = int64(0)
	identifier = int64(0)
	for i, r := range msg {
		_ = i
		msg_sum += int64(r)
	}

	identifier = int64(msg_sum % int64(number_of_network_nodes))
	return

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
		map_lock.Lock()
		network[int64(id)] = make(chan string, 100)
		map_lock.Unlock()
		//start up node
		id_64 := int64(id)
		wg.Add(1)
		go net_node(id_64)
	}
}

/*
The predecessor_id is the id of the node that is the predecesor of the node_obj
*/
func Notify(node_obj *node.Node, predecessor_id int64){


	//If the predecessor id is not in the ring, then this is a problem
	 if val, ok := ring_nodes[predecessor_id]; ok != true {
		log.Printf("\nCannot add %d as a predecessor to %d; %d is not in the ring\n", predecessor_id, node_obj.Predecessor, predecessor_id)
		return
	}
	//If node_obj already has a predecessor check to see if the predecessor_id is even closer to the node_obj.ChannelId
	//Than the existing node_obj's Predecessor
	if node_obj.Predecessor != nil {
		if predecessor_id > node_obj.Predecessor.ChannelId && predecessor_id < node_obj.ChannelId {
			node_obj.Predecessor = ring_nodes[predecessor_id]
		}

	// If node_obj does not have a predecessor yet, then assign predecessor_id as node_objs PRedecessor
	// As long as predecessor_id < node_obj.ChannelId
	}else if node_obj.Predecessor == nil {
		if predecessor_id < node_obj.ChannelId {
			node_obj.Predecessor = ring_nodes[predecessor_id]
		}else{
			log.Printf("\nPredecessor: %d is greater than Node %d. %d must be < %d", predecessor_id, node_obj.Predecessor, predecessor_id, node_obj.Predecessor)
		}
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
	//Initializing finger and datatable
	var node_obj = node.Node {ChannelId: channel_id,
				   Successor:nil,
				   Predecessor:nil,
				   FingerTable:make(map[int64]*node.Node),
				   DataTable:make(map[string]string)}

	var wait_time = int(responsetime.GetResponseTime(mean_wait_value))
	//Initialize table to size 64
	init_ring_fingers.Init_Ring_Fingers(&node_obj)
	//If ring is empty just add this node to the ring
	//This is the first node to enter the ring. Make this node's successor itself.
	//create
	if len(ring_nodes) == 0 {
		node_obj.Successor = &node_obj
		map_lock.Lock()
		ring_nodes[channel_id] = &node_obj
		map_lock.Unlock()
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

						map_lock.Lock()
						ring_nodes[channel_id] = &node_obj
						map_lock.Unlock()
					}else{
						log.Printf("\nNode %d is already in the ring; cannot join-ring\n", channel_id)
					}
			   	} else if message.Do == "leave-ring" {
					if val, ok := ring_nodes[channel_id]; ok{
						_ = val
						leave.Leave_ring(&node_obj, message.Mode)
						map_lock.Lock()
						delete(ring_nodes, channel_id)
						map_lock.Unlock()
					}else{
						log.Printf("\nNode %d is not in the ring; cannot leave-ring\n", channel_id)
					}
				} else if message.Do == "ring-notify"{

					Notify(&node_obj, message.RespondTo)
				}
				/*else if (message.Do == "put"){
					respond_to_node_id = struct_message.RespondTo
					data  = struct_message.Data
					put(data, respond_to_node_id, node_obj)
				}...
				*/
				print_ring_nodes()
				print_node(node_obj)
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
	log.Println("+++LIST OF NODES CURRENTLY IN THE RING+++")
	for channel_id, _ := range ring_nodes {
		log.Printf("\nNode %d is in the ring\n", channel_id)
	}
	log.Println("+++END OF LIST OF NODES CURRENTLY IN THE RING+++")
}

func print_node(node_obj node.Node){
log.Printf("\n+++Contents of Node %d+++\n", node_obj.ChannelId)
log.Printf("Channel Id/Node Id: %d\n", node_obj.ChannelId)
log.Printf("+FingerTable+: nil\n")
if node_obj.FingerTable != nil {
    for node_id, node_entry := range node_obj.FingerTable {
        if node_entry != nil {
		log.Printf("Finger Table entry %d is occupied\n", node_id)		
	}
    }
}

if node_obj.Successor != nil {
	log.Printf("Successor Id: %d\n", node_obj.Successor.ChannelId)
}else{
	log.Printf("Successor Id: nil\n")
}

if node_obj.Predecessor != nil {
	log.Printf("Predecessor Id: %d\n", node_obj.Predecessor.ChannelId)
}else{
	log.Printf("Predecessor Id: nil\n")
}
log.Printf("\n+++END of Contents of Node %d+++\n", node_obj.ChannelId)
	
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

	//Get a random ring nodes id
	var random_ring_id int64

	//get a list of string json instructions to send to random nodes
	var instructions []string = create_message_list(file_name)
	var channel_id int64
	for i := 0; i < len(instructions); i++ {
		//pick a random node in the ring to send the message to.
		random_ring_id = get_random_ring_node()
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

				if random_ring_id > 0 {
					message.SponsoringNode = random_ring_id
				}else{
					log.Println("There is no node to sponsor for join ring")
					continue
				}
				channel_id = random_network_id

			}else{
				channel_id = random_ring_id

			}

			modified_inst, err := json.Marshal(message)
			log.Printf("Read the following instruction from file %s.", string(modified_inst))
			check_error(err)
			// Give a random node instructions 
			network[channel_id] <- string(modified_inst) 
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


