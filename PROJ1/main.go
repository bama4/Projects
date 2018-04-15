package main

import "log"
import "os"
import "strconv"
import node "./utils/node_defs"
import msg "./utils/message_defs"
import leave "./leave_ring"
import init_ring_fingers "./init_ring_fingers"
import "math/rand"
import "sync"
import "time"
import "io/ioutil"
import "strings"
import "encoding/json"
import "math"
import responsetime "./utils/responsetime"

/*This is defining a concurent thread safe map
The ring_nodes map is NOT the chord ring
The ring nodes just allows the coordinator to know
how many nodes are in the ring for randomization purposes.

The ring nodes store node objects for each node id.
*/
type RingNodes struct {
	sync.RWMutex
	ring_nodes map[int64](*node.Node)
}

func NewRingNodesMap() *RingNodes {
	return &RingNodes{
		ring_nodes: make(map[int64](*node.Node)),
	}
}

/*Safely load a value from the map
*/
func (r_nodes *RingNodes) Load(key int64)(value *node.Node, ok bool){

	r_nodes.RLock()
	result, ok := r_nodes.ring_nodes[key]
	r_nodes.RUnlock()
	return result, ok
}


/*
Safely delete a value from the map
*/
func (r_nodes *RingNodes) Delete(key int64){
	r_nodes.Lock()
	delete(r_nodes.ring_nodes, key)
	r_nodes.Unlock()
}

/*
Safely write to the map
*/
func (r_nodes *RingNodes) Store(key int64, value *node.Node){
	r_nodes.Lock()
	r_nodes.ring_nodes[key] = value
	r_nodes.Unlock()
}

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
The map is so that the coordinator can randomly tell nodes in the ring to leave.
This is also here for debugging purposes, and so that a node object with a given 
node id/channel id can be accessed (example: update node 1's predecessor if we
are currently at node 1 in a lookup done through chord routing)

This map does not replace routing through choord, therefore no functions in this program
depend on this map for routing. Routing is done through a combination of predecessor,
successor pointer traveral and finger table lookups.
*/
var ring_nodes = NewRingNodesMap()

/*
This is a blocking map that has channels for each node in the network. This ring_node_values map can be used
for ring_nodes to recieve node objects.
*/
var ring_nodes_bucket = make(map[int64](chan *node.Node))
/*
This is the global sync group to handle the goroutines properly
*/
var wg sync.WaitGroup

/*This is a lock that should be used when writing to maps
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
        if len(ring_nodes.ring_nodes) == 0 {
            return -1
        }

        if val, ok := ring_nodes.Load(int64(rand_num)); ok {
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
		map_lock.Lock()
		ring_nodes_bucket[int64(id)] = make(chan *node.Node)
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
	 if val, ok := ring_nodes.Load(predecessor_id); ok != true {
		_ = val
		log.Printf("\nCannot add %d as a predecessor to %d; %d is not in the ring\n", predecessor_id, node_obj.Predecessor, predecessor_id)
		return
	}
	//If node_obj already has a predecessor check to see if the predecessor_id is even closer to the node_obj.ChannelId
	//Than the existing node_obj's Predecessor
	if node_obj.Predecessor != nil {
		if predecessor_id > node_obj.Predecessor.ChannelId && predecessor_id < node_obj.ChannelId {
			if val, ok := ring_nodes.Load(predecessor_id); ok {
				node_obj.Predecessor = val
			}else{
				log.Printf("\nNode %d is not in the ring\n")
			}
		}

	// If node_obj does not have a predecessor yet, then assign predecessor_id as node_objs PRedecessor
	// As long as predecessor_id < node_obj.ChannelId
	}else if node_obj.Predecessor == nil {
		if predecessor_id < node_obj.ChannelId {
			if val, ok := ring_nodes.Load(predecessor_id); ok {
				node_obj.Predecessor = val
			}else{
				log.Printf("\nNode %d is not in the ring\n")
			}
		}else{
			log.Printf("\nPredecessor: %d is greater than Node %d. %d must be < %d", predecessor_id, node_obj.Predecessor, predecessor_id, node_obj.Predecessor)
		}
	}
}

// Gets sponsoring node ID to lookup
// Node object is the node that wants to join
func Join_ring(sponsoring_node_id int64, node_obj *node.Node){
	
    node_obj.Predecessor = nil
    log.Printf("Node %d is joining the ring now", sponsoring_node_id)
    
    //The &node_obj is the node that needs a successor
    //Compose find ring successor message
    var message = msg.Message {Do:"find-ring-successor", TargetId: node_obj.ChannelId, RespondTo: sponsoring_node_id}
    string_message, err := json.Marshal(message)
    check_error(err)
	map_lock.Lock()
	network[sponsoring_node_id] <- string(string_message)
	map_lock.Unlock()
	map_lock.Lock()
	successor := <- ring_nodes_bucket[sponsoring_node_id]
	map_lock.Unlock()
	node_obj.Successor = successor
    log.Printf("\nSENT find successor message with sponsoring node: %d and target node: %d\n", sponsoring_node_id, node_obj.ChannelId)
    return
}

/*
/ ask node n to find the successor of id
n.find successor(id)
if (id ∈ (n,successor])
return successor;
else
n
0 = closest preceding node(id);
return n
0
.find successor(id);

Find the successor node of the target_id...
the sponsoring node in this case is the node_obj
the sponsoring node is sent the successor object that is found

The resulting successor is sent to the node_objs bucket entry
in the ring_nodes_bucket map
*/
func FindRingSuccessor(node_obj *node.Node, target_id int64) int {
	if node_obj.ChannelId < target_id && target_id < node_obj.Successor.ChannelId {
		log.Printf("\nFOUND a place in between for %d using find successor\n", target_id)

		//Tell node_obj that node_obj.Successor is target-ids successor (node_obj is equilvalent to respond-to)
		ring_nodes_bucket[node_obj.ChannelId] <- node_obj.Successor
		return 0

	}else if node_obj.ChannelId == node_obj.Successor.ChannelId {
		if node_obj.ChannelId > target_id {
			ring_nodes_bucket[node_obj.ChannelId] <- node_obj.Successor
			return 0
		}else{
			log.Printf("\nFOUND %d's successor is itself\n", target_id)
			if node, ok := ring_nodes.Load(target_id); ok {
				ring_nodes_bucket[node_obj.ChannelId] <- node
				node.Successor = node
			}
			return 0
		}
	}else{
		log.Printf("\nSTILL NEED TO FIND a place for %d using find successor\n", target_id)
		closest_preceeding := FindClosestPreceedingNode(node_obj, target_id)
		return FindRingSuccessor(closest_preceeding, target_id)
	}
}

/*
n.closest preceding node(id)
for i = m downto 1
if (finger[i] ∈ (n,id))
return finger[i];
return n;

respond_to is the node id that needs the closest preceeding node, node_obj is the sponsoring node
*/
func FindClosestPreceedingNode(node_obj *node.Node, respond_to int64) (closest_preceeding *node.Node){
	closest_preceeding = nil

	log.Println("Searching for closest preceeding node.....")
	for i := len(node_obj.FingerTable)-1; i >= 0; i-- {
		map_lock.Lock()
		if node_obj.FingerTable[int64(i)] != nil {
		map_lock.Unlock()
			map_lock.Lock()
			if (node_obj.FingerTable[int64(i)].ChannelId < node_obj.ChannelId && respond_to > node_obj.FingerTable[int64(i)].ChannelId) {
			map_lock.Unlock()
			map_lock.Lock()
			closest_preceeding = node_obj.FingerTable[int64(i)]
			map_lock.Unlock()
			return
			}else{
				map_lock.Unlock()			
			}
		}else{
			map_lock.Unlock()
		}
	}
	closest_preceeding = node_obj
	return
}

/*
Refreshes the finger table.
node_obj is the node that should refresh its table entries
*/
func FixRingFingers(node_obj *node.Node){

	for i :=0; i < len(node_obj.FingerTable); i++ {
		//find the successor for target id n.id + 2^i for the ith entry
		var message = msg.Message {Do:"find-ring-successor",
			TargetId: int64(node_obj.ChannelId) + int64(math.Exp2(float64(i))), RespondTo: node_obj.ChannelId}
		string_message, err := json.Marshal(message)
		check_error(err)
		map_lock.Lock()
		network[node_obj.ChannelId] <- string(string_message)
		map_lock.Unlock()

		//wait to recieve the successor result from find successor
		map_lock.Lock()
		entry_successor := <- ring_nodes_bucket[node_obj.ChannelId]
		map_lock.Unlock()
		map_lock.Lock()
		node_obj.FingerTable[int64(i)] = entry_successor
		map_lock.Unlock()
	}
	log.Printf("\nNode %d updated to the following: \n")
	print_node(node_obj)
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
	//Initialize table to size N where 2^N is the number of nodes
	init_ring_fingers.Init_Ring_FingerTable(&node_obj, int(math.Log2(float64(number_of_network_nodes))))
	//If ring is empty just add this node to the ring
	//This is the first node to enter the ring. Make this node's successor itself.
	//create
	if len(ring_nodes.ring_nodes) == 0 {
		node_obj.Successor = &node_obj
		ring_nodes.Store(channel_id, &node_obj)
		//Initialize all of the entries in the fingertable of the create node to the node
		//itself
		for i := 0; i < len(node_obj.FingerTable); i++ {
			node_obj.FingerTable[int64(i)] = &node_obj
		}
		log.Printf("Node %d was used to create the ring.", channel_id)
	}

	for {
		select {
			case msg_recv := <-network[channel_id]:
				//msg_recv := <-network[channel_id]
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
					if val, ok := ring_nodes.Load(channel_id); ok != true {
						_ = val
						sponsoring_node_id := message.SponsoringNode
						ring_nodes.Store(channel_id, &node_obj)
						Join_ring(sponsoring_node_id, &node_obj)
					}else{
						log.Printf("\nNode %d is already in the ring; cannot join-ring\n", channel_id)
					}
			   	} else if message.Do == "leave-ring" {
					if val, ok := ring_nodes.Load(channel_id); ok{
						_ = val
						leave.Leave_ring(&node_obj, message.Mode)
						ring_nodes.Delete(channel_id)
					}else{
						log.Printf("\nNode %d is not in the ring; cannot leave-ring\n", channel_id)
					}
				} else if message.Do == "ring-notify" {

					Notify(&node_obj, message.RespondTo)

				} else if message.Do == "find-ring-successor" {
					//respond-to contains the "sponsor" of this request
					if sponsor_node, ok := ring_nodes.Load(message.RespondTo); ok{
						FindRingSuccessor(sponsor_node, message.TargetId)
					}else{
						log.Printf("\nRespondTo node: %d does not exist in the ring\n", message.RespondTo)
					}
				}else if message.Do == "fix-ring-fingers"{
					FixRingFingers(&node_obj)
				}

				/*else if (message.Do == "put"){
					respond_to_node_id = struct_message.RespondTo
					data  = struct_message.Data
					put(data, respond_to_node_id, node_obj)
				}...
				*/
				print_ring_nodes()
			default:
				time.Sleep(1)
		}
	}

	log.Printf("\nNode %d left UNORDERLY\n", node_obj.ChannelId)
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
	for channel_id, node := range ring_nodes.ring_nodes {
		log.Printf("\nNode %d is in the ring\n", channel_id)
		print_node(node)
	}
	log.Println("+++END OF LIST OF NODES CURRENTLY IN THE RING+++")
}

func print_node(node_obj *node.Node){
log.Printf("\n+++Contents of Node %d+++\n", node_obj.ChannelId)
log.Printf("Channel Id/Node Id: %d\n", node_obj.ChannelId)
log.Printf("+FingerTable+:\n")
if node_obj.FingerTable != nil {
    for node_id, node_entry := range node_obj.FingerTable {
        if node_entry != nil {
		log.Printf("Finger Table at %d is %d\n", node_id, node_entry.ChannelId)		
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

	//The number of nodes is 2^N where N is what the user entered for the amount of nodes
	number_of_network_nodes = int(math.Exp2(float64(num_nodes)))
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

			}else if message.Do == "fix-ring-fingers" {
				channel_id = random_ring_id
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
		log.Println("USAGE: go run main.go <INSTRUCTION FILE> <N WHERE #NODES is 2^N> <AVERAGE_WEIGHT_TIME>")
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


