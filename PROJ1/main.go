package main

import "log"
import "os"
import "strconv"
import node "./utils/node_defs"
import msg "./utils/message_defs"
//import leave "./leave_ring"
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

The ring nodes store node objects for each node id for debugging 
purposes
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
successor pointer traversal and finger table lookups.
*/
var ring_nodes = NewRingNodesMap()

/*
This is a blocking map that has channels for each node in the network. This ring_node_values map can be used
for nodes in the ring to recieve node objects.
*/
var ring_nodes_bucket = make(map[int64](chan int64))
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

/*Gets a random node in the chord ring so that the coordinator can send
instructions such as get/put/ and leave to random nodes in the ring
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
Generates a random integer
*/
func get_random_int()(rand_num int){

	rand_num = rand.Intn(1000)
	return
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
		ring_nodes_bucket[int64(id)] = make(chan int64, 5)
		map_lock.Unlock()
		//start up node
		id_64 := int64(id)
		wg.Add(1)
		go net_node(id_64)
		if i == 0 {
			//Wait a little while for the first node to be created and start the ring
			time.Sleep(3)
		}
	}
}

/*
The predecessor is the id of the node that is the predecesor of the node_obj
*/
func Notify(node_obj *node.Node, predecessor int64){

	//If node_obj already has a predecessor check to see if the predecessor is even closer to the node_obj.ChannelId
	//Than the existing node_obj's Predecessor
	if node_obj.Predecessor != -1 {
		if predecessor > node_obj.Predecessor && predecessor < node_obj.ChannelId {
				node_obj.Predecessor = predecessor
		}

	// If node_obj does not have a predecessor yet, then assign predecessor as node_objs PRedecessor
	// As long as predecessor < node_obj.ChannelId
	}else if node_obj.Predecessor == -1 {
		if predecessor < node_obj.ChannelId {
				node_obj.Predecessor = predecessor
		}else{
		
	log.Printf("\nPredecessor: %d is greater than Node %d. %d must be < %d", predecessor, node_obj.Predecessor, predecessor, node_obj.Predecessor)
		}
	}
}

func GetNodeRoutineObj(node_id int64)(ring_node *node.Node){
	ring_node = ring_nodes.ring_nodes[node_id]
	return
}

/*
This function sends data to a node_id's bucket
*/
func SendDataToBucket(node_id int64, bucket_data int64){
	log.Printf("\nBUCKET:Node: %d was written bucket data\n", node_id)
	map_lock.Lock()
	ring_nodes_bucket[node_id] <- bucket_data
	map_lock.Unlock()
	return
}
/*
This function recieves data from the designated bucket.
The node id given is used to read the correct bucket
*/
func GetDataFromBucket(node_id int64)(bucket_data int64){
	log.Printf("\nBUCKET:Node: %d's  data is being read ....\n", node_id)
	bucket_data = <- ring_nodes_bucket[node_id]
	log.Printf("\nBUCKET:Node: %d's data has finished being read ....\n", node_id)
	return
}

/*
This function sends the data through the network.
The node id is the identifier used to get the right channel to send on
*/
func SendDataToNetwork(node_id int64, msg string){
	map_lock.Lock()
	network[node_id] <- string(msg)
	map_lock.Unlock()
}

/*
Retrieves the idx index of the given nodes finger table
*/
func ReadNodeFingerTable(node_obj *node.Node, idx int64)(result int64){
	map_lock.Lock()
	result = node_obj.FingerTable[idx]
	map_lock.Unlock()
	return
}

/*
 Gets sponsoring node ID to lookup
 Node object is the node that wants to join
 Sends the sponsoring node a find-ring-successor message through the network
 waits to get back the message representing the successor node
*/

func Leave_ring(leave_node *node.Node, mode string) {

        switch mode {
                case "immediate":
                        leave_node.Predecessor = -1
                        leave_node.Successor = -1
                        leave_node.FingerTable = nil
                        log.Printf("\nNode: %d is leaving immediately\n", leave_node.ChannelId)

                case "orderly":
                        log.Printf("\nNode: %d is leaving orderly\n", leave_node.ChannelId)

                        // Notify Successor and pRedecessor we are leaving
                        var leaving_msg = msg.Message {Do:"leaving", TargetId:leave_node.ChannelId}
                        //var leave_succ = msg.Message {Do:"set-successor", TargetId:leave_node.ChannelID }
                        var leave_pred = msg.Message {Do:"set-predecessor", TargetId:leave_node.Successor }

						var json_leave, _ = json.Marshal(leaving_msg)
						var json_pred, _ = json.Marshal(leave_pred)

                        SendDataToNetwork(leave_node.Successor, string(json_leave))
                        SendDataToNetwork(leave_node.Predecessor, string(json_leave))
                        //SendDataToNetwork(leave_node.Predecessor, leave_succ)
                        SendDataToNetwork(leave_node.Successor, string(json_pred))

                        // Loop through current nodes DataTable to append to successor
                        for k, v := range leave_node.DataTable {

								// Somehow fix Data

                                var message msg.Message = {Do:"store-data-successor", Data:{k:v} }
                                string_msg, _ := json.Marshal(message)

                                // Send Data to Successor
                                SendDataToNetwork(leave_node.ChannelId, string_message)
                        }

                        // remove node from ring
                        leave_node.Predecessor = -1
                        leave_node.Successor = -1
                        leave_node.FingerTable = nil

                default:
                        // Immediate leave
                       leave_node.Predecessor = -1
                        leave_node.Successor = -1
                        leave_node.FingerTable = nil
                        log.Printf("\nNode: %d is leaving immediately\n", node.ChannelId)
        }

}

func Join_ring(sponsoring_node_id int64, node_obj *node.Node){
	
    node_obj.Predecessor = -1
    log.Printf("\nJOIN_RING:Node %d is joining the ring now\n", node_obj.ChannelId)
    
    //The &node_obj is the node that needs a successor
    //Compose find ring successor message
    var message = msg.Message {Do:"find-ring-successor", TargetId: node_obj.ChannelId, RespondTo: sponsoring_node_id}

    string_message, err := json.Marshal(message)
    check_error(err)
	//Tell sponsoring_node_id to find the successor of the node_obj
	SendDataToNetwork(sponsoring_node_id, string(string_message))
	//Wait to hear back what the successor is
	successor := GetDataFromBucket(sponsoring_node_id)
	if successor != -1 {
		node_obj.Successor = successor
	}else{
		node_obj.Successor = node_obj.ChannelId
	}
    log.Printf("\nJOIN_RING:SENT find successor message with sponsoring node: %d and target node: %d\n", sponsoring_node_id, node_obj.ChannelId)
    return
}


/*
node_obj is the initial potential 
target id is the node id that is looking for a predecessor
respond-to is the id of the node that was asked to find the predecessor of targetid
*/
func FindRingPredecessor(node_obj *node.Node, target_id int64, respond_to int64){
	potential_predecessor := node_obj

	log.Printf("\nSearching for predecessor for %d with sponsoring node as %d\n", target_id, node_obj.ChannelId)
	// While the target id is not between the 
	for !(target_id > potential_predecessor.ChannelId && target_id < potential_predecessor.Successor){
		FindClosestPreceedingNode(potential_predecessor, target_id) //find the closest preceeding node from the target-id
		potential_predecessor_id := GetDataFromBucket(node_obj.ChannelId)
		potential_predecessor = GetNodeRoutineObj(potential_predecessor_id)
		//If the potential predecessor is equal to the node
		//that sponsored finding the predecessor...
		if potential_predecessor.ChannelId == node_obj.ChannelId{
			if node_obj.ChannelId < target_id {
				//Tell the node_obj (respond to node) that
				//the predecessor of target_id is node_obj
				SendDataToBucket(respond_to, node_obj.ChannelId)
				return
			}
		}
	}
	SendDataToBucket(node_obj.ChannelId, node_obj.ChannelId)
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

The resulting successor is sent to the goroutine nodes bucket entry
in the ring_nodes_bucket map
*/
func FindRingSuccessor(node_obj *node.Node, target_id int64, respond_to int64) int {
	log.Printf("\nFIND_SUCCESSOR:Finding the successor of %d by asking Node: %d\n", target_id, node_obj.ChannelId)
	if node_obj.ChannelId < target_id && target_id < node_obj.Successor {
		log.Printf("\nFIND_SUCCESSOR:FOUND a place in between for %d using find successor\n", target_id)

		//Tell node_obj that node_obj.Successor is target-ids successor (node_obj is equilvalent to respond-to)
		SendDataToBucket(respond_to, node_obj.Successor)
		return 0

	}else if node_obj.ChannelId == node_obj.Successor {

		//Tell the respond-to that the successor is the nodes successor
		SendDataToBucket(respond_to, node_obj.Successor)
		return 0

	}else{
		log.Printf("\nFIND_SUCCESSOR:STILL NEED TO FIND a successor for %d and tell %d\n", target_id, respond_to)
		// var message = msg.Message {Do:"find-closest-preceeding-node", TargetId: target_id, RespondTo: node_obj.ChannelId}
    		//string_message, err := json.Marshal(message)
    		//check_error(err)
		//Tell the sponsoring node_obj to Find the closest preceeding node of target_id
		//SendDataToNetwork(node_obj.ChannelId, string(string_message))
		FindClosestPreceedingNode(node_obj, target_id)
		closest_preceeding := GetDataFromBucket(node_obj.ChannelId)
		log.Printf("\nFIND_SUCCESSOR: Node %d Found the closest preceeding node of %d to be %d\n", node_obj.ChannelId, target_id, closest_preceeding)
		next_successor := GetNodeRoutineObj(closest_preceeding)

		//If the closest preceeding node is the node that initiated the request..then just return the nodes successor
		if closest_preceeding == node_obj.ChannelId {
			SendDataToBucket(respond_to, node_obj.Successor)
			return 0
		}
		return FindRingSuccessor(next_successor, target_id, respond_to)
	}
}

/*
n.closest preceding node(id)
for i = m downto 1
if (finger[i] ∈ (n,id))
return finger[i];
return n;
corresponding json message is {"do":"find-closest-preceeding-node", "target-id": target, "respond-to": respond}
respond_to is the node_obj that needs to find the closest preceeding node in its table to targetid
*/
func FindClosestPreceedingNode(node_obj *node.Node, target_id int64){
	var closest_preceeding int64 = node_obj.ChannelId

	log.Println("CLOSEST_PRECEEDING:Searching for closest preceeding node.....")
	for i := len(node_obj.FingerTable)-1; i >= 0; i-- {
		finger_entry := ReadNodeFingerTable(node_obj, int64(i))
		if finger_entry != -1 {
			if (finger_entry < node_obj.ChannelId && target_id > finger_entry) {
			
			closest_preceeding := ReadNodeFingerTable(node_obj, int64(i))
			SendDataToBucket(node_obj.ChannelId, closest_preceeding)
			return
			}
		}
		
	}
	SendDataToBucket(node_obj.ChannelId, closest_preceeding)
	return
}

/*
Refreshes the finger table.
node_obj is the node that should refresh its table entries
*/
func FixRingFingers(node_obj *node.Node){

	for i :=0; i < len(node_obj.FingerTable); i++ {
		//find the successor for target id n.id + 2^i for the ith entry
		FindRingSuccessor(node_obj, int64(node_obj.ChannelId) + int64(math.Exp2(float64(i))), node_obj.ChannelId)
		log.Printf("\nFIX_FINGERS:Looking for %d's successor at entry %d for Node %d\n", 
			int64(node_obj.ChannelId) + int64(math.Exp2(float64(i))), i, node_obj.ChannelId)
		//wait to recieve the successor result from find successor
		entry_successor := GetDataFromBucket(node_obj.ChannelId)
		log.Printf("\nFIX_FINGERS:Recieved successor %d for entry %d\n", entry_successor, i)
		map_lock.Lock()
		node_obj.FingerTable[int64(i)] = entry_successor
		map_lock.Unlock()
	}
	log.Printf("\nFIX_FINGERS:Node %d updated to the following: \n")
	print_node(node_obj)
}


/*
This function removes data from the chord ring.

func RemoveData(node_obj *node.Node, data Data, respond_to int64){
	id = map_string_to_int(data.Key)
	FindClosestPrecedingNode()
}
*/

/*
This is a routine that defines a node. The routine listens on the channel that is assigned
to the given channel id  for incoming messages. The routine stores its pointer information
and table in a node message structure
*/
func net_node(channel_id int64){
	
        defer wg.Done()
	//create a node structure to store information,
	//successor/predecessor references, etc.
	//Initializing finger and datatable
	var node_obj = node.Node {ChannelId: channel_id,
				   Successor: -1,
				   Predecessor: -1,
				   FingerTable:make(map[int64]int64),
				   DataTable:make(map[string]string)}

	var wait_time = int(responsetime.GetResponseTime(mean_wait_value))
	//Initialize table to size N where 2^N is the number of nodes
	init_ring_fingers.Init_Ring_FingerTable(&node_obj, int(math.Log2(float64(number_of_network_nodes))))
	//If ring is empty just add this node to the ring
	//This is the first node to enter the ring. Make this node's successor itself.
	//create
	if len(ring_nodes.ring_nodes) == 0 {
		node_obj.Successor = node_obj.ChannelId
		ring_nodes.Store(channel_id, &node_obj)
		//Initialize all of the entries in the fingertable of the create node to the node
		//itself
		for i := 0; i < len(node_obj.FingerTable); i++ {
			map_lock.Lock()
			node_obj.FingerTable[int64(i)] = node_obj.ChannelId
			map_lock.Unlock()
		}
		log.Printf("Node %d was used to create the ring.", channel_id)
	}

	for {
		select {
			//This node recieved a message
			case msg_recv := <-network[channel_id]:
				//wait an average of AVERAGE_WAIT_TIME seconds before accepting a message
				log.Printf("\nWaiting %d seconds before processing message for Node: %d\n", wait_time, channel_id)
				wait_time = int(responsetime.GetResponseTime(mean_wait_value))
				time.Sleep(time.Duration(wait_time) * time.Second)
				log.Printf("\nNode: %d received the following message:%s\n", channel_id, msg_recv)

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
						Join_ring(sponsoring_node_id, &node_obj)
						ring_nodes.Store(channel_id, &node_obj)
						execute_fix_fingers := get_random_int() % 2 == 0
						//Randomly tell the joining node to fix its fingers
						if execute_fix_fingers == true {
							SendDataToNetwork(node_obj.ChannelId, "{\"do\": \"fix-ring-fingers\"}")
						}
					}else{
						log.Printf("\nNode %d is already in the ring; cannot join-ring\n", channel_id)
					}
			   	} else if message.Do == "leave-ring" {
					if val, ok := ring_nodes.Load(channel_id); ok{
						_ = val
						Leave_ring(&node_obj, message.Mode)
						ring_nodes.Delete(channel_id)
					}else{
						log.Printf("\nNode %d is not in the ring; cannot leave-ring\n", channel_id)
					}
				} else if message.Do == "ring-notify" {
					Notify(&node_obj, message.RespondTo)
				} else if message.Do == "find-ring-successor" {
					//respond-to contains the "sponsor" of this request
					//respond-to is the node that recieves the answer of find ring successor
					if sponsor_node, ok := ring_nodes.Load(message.RespondTo); ok{
						FindRingSuccessor(sponsor_node, message.TargetId, message.RespondTo)
						execute_fix_fingers := get_random_int() % 2 == 0
						if execute_fix_fingers == true {
							SendDataToNetwork(node_obj.ChannelId, "{\"do\": \"fix-ring-fingers\"}")
						}
					}else{
						log.Printf("\nRespondTo node: %d is not responding...not in ring?\n", message.RespondTo)

					}
				} else if message.Do == "find-ring-predecessor" {
					//Tell node_obj to find the predecessor of target id and report back to respond-to
					FindRingPredecessor(&node_obj, message.TargetId, message.RespondTo)
				} else if message.Do == "store-data-successor" {
					// Store the data to a nodes successor data table
					node_obj.DataTable[string(message.TargetId)] = string(message.Data)
					
				} else if message.Do == "fix-ring-fingers"{
					FixRingFingers(&node_obj)

				}else if message.Do == "find-closest-preceeding-node" {
					//Have node_obj find the closest preceeding node to the target_id
					FindClosestPreceedingNode(&node_obj, message.TargetId)

				//Tell node to set its successor to target-id
				}else if message.Do == "set-successor" {
					//Set the successor as the target id
					node_obj.Successor = message.TargetId

				} else if message.Do == "set-predecessor" {
					// Set the predecessor to the target ID
					node.obj.Predecessor = message.TargetId
				}

				/*else if message.Do == "put" {
					respond_to_node_id = struct_message.RespondTo
					data  = struct_message.Data
					Put(data, respond_to_node_id, node_obj)
				}else if message.Do == "get" {
					respond_to_node_id = struct_message.RespondTo
					data  = struct_message.Data
					Get(data, respond_to_node_id, node_obj)

				}else if message.Do == "remove"{
					respond_to_node_id = struct_message.RespondTo
					data  = struct_message.Data
					Remove(data, respond_to_node_id, node_obj)
				}
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

/*
Prints the ring nodes for debugging purposes
*/
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
        if node_entry != -1 {
		log.Printf("Finger Table at %d is %d\n", node_id, node_entry)		
	}
    }
}
log.Printf("+DataTable+:\n")
if node_obj.DataTable != nil {
    for key, value := range node_obj.DataTable {
        log.Printf("Data Table at %s is %s\n", key, value)
    }
}

if node_obj.Successor != -1 {
	log.Printf("Successor Id: %d\n", node_obj.Successor)
}else{
	log.Printf("Successor Id: nil\n")
}

if node_obj.Predecessor != -1 {
	log.Printf("Predecessor Id: %d\n", node_obj.Predecessor)
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

/*
This is the coordinator that reads join/leave/put/get/remove instructions from a file
The coordinator randomly generates the sponsoring node for join ring
*/
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
				if channel_id < 0 {
					log.Println("There is no node in the ring to fix fingers")
					continue
				}
			}else{
				channel_id = random_ring_id

			}

			modified_inst, err := json.Marshal(message)
			check_error(err)
			// Give a random node instructions 
			network[channel_id] <- string(modified_inst)
	}
	
}


/*
This is the main function which takes in the parameters for the program.
The parameters are the instruction file with each line containing
json-formatted instruction messages, the number of nodes to generate for the network, 
and the mean variable to use in the randomization of the node response times.
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


