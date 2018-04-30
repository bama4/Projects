package main

import "log"
import "os"
import "strconv"
import node "./utils/node_defs"
import msg "./utils/message_defs"
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
var ring_nodes_bucket = make(map[int64](chan string))
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
var number_of_network_nodes int = 4

/*
This is the global test configuration
*/
var test_mode bool = false

/*
This is the first node in the ring in test mode
*/
var test_first_node int64 = 2

/*
This is the test channel
*/
var test_channel = make(chan string)

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
    defer map_lock.Unlock()

    map_lock.Lock()
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
    defer map_lock.Unlock()
    rand_num = 0
    if len(network) == number_of_network_nodes {
        map_lock.Lock()
        //cant generate a unique id
        return -1
    }

    map_lock.Lock()
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
    
    //Adds the designated first node for test to the network
    if test_mode == true {
        AddNodeToNetwork(test_first_node)
        //start up node
        wg.Add(1)
        go net_node(test_first_node)
        //Wait for first node to be set up
        _ = <-test_channel
    }

    for i:=1; i <= number_of_network_nodes; i++ {
        id := generate_channel_id()
        //Return if failed to generate an id that is not already in the network
        if id == -1{
            log.Println("INIT_TOPOLOGY:Failed to generate any more network nodes")
            return
        }
        log.Printf("\nAdding Node %d to network\n", id)
        //add node to network
        AddNodeToNetwork(int64(id))
        //start up node
        id_64 := int64(id)
        wg.Add(1)
        go net_node(id_64)
    }
}

/*
Adds a node to the network setting up its network channel and
bucket channel
*/
func AddNodeToNetwork(id int64){
    //add node to network
    map_lock.Lock()
    network[int64(id)] = make(chan string, 100)
    map_lock.Unlock()
    map_lock.Lock()
    ring_nodes_bucket[int64(id)] = make(chan string, 5)
    map_lock.Unlock()
}
/*
The predecessor is the id of the node that is the predecesor of the node_obj
*/
func Notify(node_obj *node.Node, predecessor int64){

    //If node_obj already has a predecessor check to see if the predecessor is even closer to the node_obj.ChannelId
    //Than the existing node_obj's Predecessor
    if node_obj.Predecessor == -1 || Between(predecessor, node_obj.Predecessor, node_obj.ChannelId){
                node_obj.Predecessor = predecessor
        }
}

func GetNodeRoutineObj(node_id int64)(ring_node *node.Node){
    ring_node = ring_nodes.ring_nodes[node_id]
    return
}

/*
This function sends data to a node_id's bucket
The message tag ensures that the order in which
the nodes gets the information is correct
*/
func SendDataToBucket(node_id int64, bucket_data string){
    log.Printf("\nBUCKET:Node: %d was written bucket data\n", node_id)
    map_lock.Lock()
    ring_nodes_bucket[node_id] <- bucket_data
    map_lock.Unlock()
    return
}

/*
Returns the identifier field as int64 from a given BucketMessage formatted json string
*/
func ExtractIdFromBucketData(data string)(identifier int64){
    byte_msg := []byte(data)
    var message msg.BucketMessage
    err := json.Unmarshal(byte_msg, &message)
    check_error(err)
    identifier = message.Identifier
    return
}
/*
This function recieves data from the designated bucket.
The node id given is used to read the correct bucket
*/
func GetDataFromBucket(node_id int64)(bucket_data string){
    log.Printf("\nBUCKET:Node: %d's  data is waiting for data to be read ....\n", node_id)
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
    bucket_data := GetDataFromBucket(sponsoring_node_id)
    successor := ExtractIdFromBucketData(bucket_data)
    if successor != -1 {
        node_obj.Successor = successor
    }else{
        node_obj.Successor = node_obj.ChannelId
    }
    log.Printf("\nJOIN_RING:SENT find successor message with sponsoring node: %d and target node: %d. Successor of target is %d\n", sponsoring_node_id, node_obj.ChannelId, successor)
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
        bucket_data := GetDataFromBucket(node_obj.ChannelId)
        potential_predecessor_id := ExtractIdFromBucketData(bucket_data)
        //If the potential predecessor is equal to the node
        //that sponsored finding the predecessor...
        if potential_predecessor_id == node_obj.ChannelId{
            if node_obj.ChannelId < target_id {
                //Tell the node_obj (respond to node) that
                //the predecessor of target_id is node_obj
                var bucket_msg =  msg.BucketMessage {Identifier: node_obj.ChannelId}
                string_message, err := json.Marshal(bucket_msg)
                    check_error(err)
                SendDataToBucket(respond_to, string(string_message))
                return
            }
        }
    }
    var bucket_msg =  msg.BucketMessage {Identifier: node_obj.ChannelId}
    string_message, err := json.Marshal(bucket_msg)
    check_error(err)
    SendDataToBucket(node_obj.ChannelId, string(string_message))
    return 
}


/*
This function stabilizes the ring
it sends a {do: get-predecessor respond-to: node_obj.ChannelId} to the node_obj.Successor
to tell the node_obj what the node_objs.Successor's Predecessor is
*/
func Stabilize(node_obj *node.Node){

    log.Printf("\nSTABILIZE:About to stabilize Node %d\n", node_obj.ChannelId)
    var x int64 = -1
    //If node_objs successor is itself, we can just get the 
    //predecessor directly
    //Send a message that you are looking for the
    //predecessor of node_obj.Successor to see if node_obj.Successor.Predecessor
    //Should instead be node_obj's.Sucessor
    var message = msg.Message {Do:"get-predecessor", RespondTo: node_obj.ChannelId}
    string_message, err := json.Marshal(message)
    log.Printf("\nSTABILIZE: To Stabilize Node %d, told %d to return predecessor\n", node_obj.ChannelId, node_obj.Successor)
    //Only send a get predecessor message if the node_obj and its successor are two different nodes
    if node_obj.ChannelId != node_obj.Successor {
        var message = msg.Message {Do:"get-predecessor", RespondTo: node_obj.ChannelId}
        string_message, err := json.Marshal(message)
        check_error(err)
        SendDataToNetwork(node_obj.Successor, string(string_message))
        //Listen for the response containing the predecessor id
        bucket_data := GetDataFromBucket(node_obj.ChannelId)
        x = ExtractIdFromBucketData(bucket_data)
    }else {
        x = node_obj.Predecessor
    }

    if Between(x, node_obj.ChannelId, node_obj.Successor) && x != -1{
        //Set node_objs Successor to x
        node_obj.Successor = x
        log.Printf("\nSTABILIZE: Node %d's successor has been set to Node %d\n", node_obj.ChannelId, node_obj.Successor)
    }

    //Tell node_obj.Successor that node_obj may be the predecessor
    log.Printf("\nSTABILIZE: Node %d is Telling Node %d to perform Notify\n", node_obj.ChannelId, node_obj.Successor)
    message = msg.Message {Do:"ring-notify", RespondTo: node_obj.ChannelId}
    string_message, err = json.Marshal(message)
    check_error(err)
    SendDataToNetwork(node_obj.Successor, string(string_message))

}


/*
Gets the data...The node_obj is the node that will start the lookup for the data
The data that is obtained will be sent through the bucket
*/
func GetData(node_obj *node.Node, respond_to int64, key string){
    log.Printf("\nGetting data with key %s by asking Node %d\n", key, node_obj.ChannelId)
    key_id := map_string_to_id(key)
    log.Printf("\nKey: %s mapped to hash of %d\n", key, key_id)
    FindClosestPreceedingNode(node_obj, key_id)
    bucket_data := GetDataFromBucket(node_obj.ChannelId)
    closest := ExtractIdFromBucketData(bucket_data)
    log.Printf("\nGET: Found %d as the closest to %d\n", closest, key_id)
    if closest > key_id {
        //Then just say we are at the right node to store
        log.Printf("\nStored Data\n")
        
    }
    return
}

/*
Removes the data...The node_obj is the node that will start the lookup for the data
The data that is obtained will be sent through the bucket
*/
func RemoveData(node_obj *node.Node, respond_to int64, key string){
    log.Printf("\nGetting data with key %s by asking Node %d\n", key, node_obj.ChannelId)
    key_id := map_string_to_id(key)
    log.Printf("\nKey: %s mapped to hash of %d\n", key, key_id)
    FindClosestPreceedingNode(node_obj, key_id)
    bucket_data := GetDataFromBucket(node_obj.ChannelId)
    closest := ExtractIdFromBucketData(bucket_data)
    log.Printf("\nREMOVE: Found %d as the closest to %d\n", closest, key_id)

    if closest > key_id {
        //Then just say we are at the right node to store
        log.Printf("\nStored Data\n")
        
    }
    return
}

func PutData(node_obj *node.Node, respond_to int64, key string, value string) {

    log.Printf("\nPutting data with key %s by asking Node %d\n", key, node_obj.ChannelId)
    key_id := map_string_to_id(key)
    log.Printf("\nKey: %s mapped to hash of %d\n", key, key_id)
    FindClosestPreceedingNode(node_obj, key_id)
    bucket_data := GetDataFromBucket(node_obj.ChannelId)
    closest := ExtractIdFromBucketData(bucket_data)

    log.Printf("\nPUT: Found %d as the closest to %d\n", closest, key_id)

    if closest > key_id {
        //Then just say we are at the right node to store
		log.Printf("\nPUT: Putting Key: %s with value: %s at Node: %d\n", key, value, node_obj.ChannelId) 
		node_obj.DataTable[key] = value
        
    } else {
		// This is the wrong node to store the data
		// Need to send this message to the successor
		// Get Successor Channel ID
		// Send Message to Successor
		//node_obj.DataTable[key] := value

	 	log.Printf("\nPUT: Sending key: %s and value: %s to bucket for Node: %d\n", key, value, node_obj.Successor)	
		SendDataToBucket(node_obj.Successor, key)
		SendDataToBucket(node_obj.Successor, value)
	}

    return
}


/*
A function that compares the target_id relative to the ordering of the first
and second id
*/
func Between(target_id int64, first int64, second int64)(result bool){
    log.Printf("\nBETWEEN: Checking if %d is in between %d and %d\n", target_id, first,second)
    result = true
    //The first and second node is in order
    //So you can return true if the target_id is in between
    if first < second {
        if first < target_id && target_id < second{
            result = true
            return
        }else{
            result = false
            return
        }
    
    //The ring is not in order
    //So you can return true only if the target comes after the first or before the second
    }else {
        log.Printf("BEETWEEN: %d target and %d first\n", target_id, first)
        return target_id > first || target_id < second
    }
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
    //check if either target_id is in between the node and node.successor if node/node.successor is in order
    //or if they are not in order make sure that the target_id is either greater than the node or less than the node.successor
    if Between(target_id, node_obj.ChannelId, node_obj.Successor) || target_id == node_obj.Successor{
        log.Printf("\nFIND_SUCCESSOR:FOUND a place in between for %d. Successor is %d\n", target_id, node_obj.Successor)

        //Tell node_obj that node_obj.Successor is target-ids successor (node_obj is equilvalent to respond-to)
        var bucket_msg =  msg.BucketMessage {Identifier: node_obj.Successor}
        string_message, err := json.Marshal(bucket_msg)
        check_error(err)
        SendDataToBucket(respond_to, string(string_message))
        return 0

    //This is the case where the target_id is not in between node and node.successor and the target_id is less 
    //than the node but greater than the nodes successor
    }else{
        log.Printf("\nFIND_SUCCESSOR: Node %d is not between %d and %d\n", target_id, node_obj.ChannelId, node_obj.Successor)
        log.Printf("\nFIND_SUCCESSOR:STILL NEED TO FIND a successor for %d and tell %d...will look at %d's table\n", target_id, respond_to, node_obj.ChannelId)
        // var message = msg.Message {Do:"find-closest-preceeding-node", TargetId: target_id, RespondTo: node_obj.ChannelId}
            //string_message, err := json.Marshal(message)
            //check_error(err)
        //Tell the sponsoring node_obj to Find the closest preceeding node of target_id
        //SendDataToNetwork(node_obj.ChannelId, string(string_message))
        FindClosestPreceedingNode(node_obj, target_id)
        bucket_data := GetDataFromBucket(node_obj.ChannelId)
        closest_preceeding := ExtractIdFromBucketData(bucket_data)
        log.Printf("\nFIND_SUCCESSOR: Node %d Found the closest preceeding node of %d to be %d\n", node_obj.ChannelId, target_id, closest_preceeding)
        //If we are at the closest_preceeding node, then just return that as the successor
        if closest_preceeding == node_obj.ChannelId {
            var bucket_msg =  msg.BucketMessage {Identifier: node_obj.ChannelId}
            string_message, err := json.Marshal(bucket_msg)
            check_error(err)
            SendDataToBucket(respond_to, string(string_message))
            return 0
        }else{
            //Tell the closest_preceeding node to find successor
            var bucket_msg =  msg.Message {Do: "find-ring-successor", TargetId: target_id, RespondTo: respond_to}
            string_message, err := json.Marshal(bucket_msg)
            check_error(err)
            SendDataToNetwork(closest_preceeding, string(string_message))
            return 0
        }
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
            //If the entry is 
            if Between(finger_entry, node_obj.ChannelId, target_id) {            
                //Send the closest preceeding id to the respond-to node that requested it
                closest_preceeding := finger_entry
                var bucket_msg =  msg.BucketMessage {Identifier: closest_preceeding}
                string_message, err := json.Marshal(bucket_msg)
                check_error(err)
                SendDataToBucket(node_obj.ChannelId, string(string_message))
                return
            }
        }
        
    }
    
    //Send closest proceeding to respond-to
    var bucket_msg =  msg.BucketMessage {Identifier: closest_preceeding}
    string_message, err := json.Marshal(bucket_msg)
    check_error(err)
    SendDataToBucket(node_obj.ChannelId, string(string_message))
    return
}

/*
Refreshes the finger table.
node_obj is the node that should refresh its table entries
*/
func FixRingFingers(node_obj *node.Node){
    var target_id int64
    var raw_entry_id int64
    for i :=0; i < len(node_obj.FingerTable); i++ {
        //find the successor for target id n.id + 2^i for the ith entry
        raw_entry_id = int64(node_obj.ChannelId) + int64(math.Exp2(float64(i)))
        target_id =  ShiftMod(raw_entry_id, int64(number_of_network_nodes))
        log.Printf("\nFIX_FINGERS:Looking for %d --> %d successor at entry %d for Node %d\n", raw_entry_id, target_id, i, node_obj.ChannelId)

        //Dont send the find successor message if we are already at the successor
        if node_obj.Successor == node_obj.ChannelId{
            FindRingSuccessor(node_obj, target_id, node_obj.ChannelId)
        }else{
            log.Printf("\nSending a find-ring-successor for %d --> %d to Node %d\n", raw_entry_id, target_id, node_obj.Successor)
            var bucket_msg =  msg.Message {Do: "find-ring-successor", TargetId: target_id, RespondTo: node_obj.ChannelId}
                string_message, err := json.Marshal(bucket_msg)
                check_error(err)
                SendDataToNetwork(node_obj.Successor, string(string_message))
        }
        //wait to recieve the successor result from find successor
        bucket_data := GetDataFromBucket(node_obj.ChannelId)
        entry_successor := ExtractIdFromBucketData(bucket_data)
        log.Printf("\nFIX_FINGERS:Recieved successor %d for %d --> %d at entry %d of Node %d's table\n", entry_successor, raw_entry_id, target_id , i, node_obj.ChannelId)
        map_lock.Lock()
        node_obj.FingerTable[int64(i)] = entry_successor
        map_lock.Unlock()
    }
    log.Printf("\nFIX_FINGERS:Node %d updated to the following: \n", node_obj.ChannelId)
    print_node(node_obj)
}

func ShiftMod(num int64, fact int64) int64{

    divide_result := num % fact
    return divide_result
}

/*
Checks to see if the predecessor node has failed
*/
func CheckPredecessor(node_obj *node.Node){
    if HasNodeFailure(node_obj.Predecessor){
        node_obj.Predecessor = -1    
    }
}

/* Temporary placeholder for a timeout through a channel
to detect node failure
*/
func HasNodeFailure(node_id int64)(bool){
    if val, ok := ring_nodes.Load(node_id); ok{
        _ = val
        return false
    }
    return true
}

/*
The given node leaves the ring.
The node notifies its successor that node.Predecessor may be its successor.
The node also tells its predecessor to set its successor to the nodes successor
*/
func Leave_ring(node *node.Node, mode string) {

    // Leaves orderly or immediate
    switch mode { 
        case "immediate":
            node.Predecessor = -1
            node.Successor = node.ChannelId
            //Clear finger table
            for k,_ := range node.FingerTable {

                node.FingerTable[k] = -1
            }

            log.Printf("\nNode: %d is leaving immediately\n", node.ChannelId)
            
        case "orderly":
            log.Printf("\nNode: %d is leaving orderly\n", node.ChannelId)
            // stuff to tell other nodes
            //Let successor know that the node is leaving
            //and to suggest node.Predecessor as new Predecessor
            var message = msg.Message {Do:"ring-notify", RespondTo: node.Predecessor}

             string_message, err := json.Marshal(message)
             check_error(err)
             SendDataToNetwork(node.Successor, string(string_message))

            //Tell predecessor to update successor
             message = msg.Message {Do:"set-successor", TargetId: node.Successor}

             string_message, err = json.Marshal(message)
             check_error(err)
             SendDataToNetwork(node.Successor, string(string_message))

             message = msg.Message {Do:"fix-fingers"}

             string_message, err = json.Marshal(message)
             check_error(err)
             SendDataToNetwork(node.Predecessor, string(string_message))
            // Loop through nodes fingertable to append to successor
    
            // remove node from ring
            //node.Predecessor = -1
            //node.Successor = -1
            
            for k,_ := range node.FingerTable {

                node.FingerTable[k] = -1
            }

        default:
            // Immediate leave
            node.Predecessor = -1
            node.Successor = node.ChannelId
            //Clear finger table
            for k,_ := range node.FingerTable {

                node.FingerTable[k] = -1
            }
            log.Printf("\nNode: %d is leaving immediately\n", node.ChannelId)
    }

}

/*
This function removes data from the chord ring.
*/
/*
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
                   Successor: channel_id,
                   Predecessor: channel_id,
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

        if test_mode == true{
            test_channel <- "Done"
        }
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

				//log.Printf("\n---- UnMarshal Data Struct: %+v -----\n", message)
                if err != nil {
                    log.Printf("Node: %d failed to unmarshal the json string", channel_id)
                    break
                }
                //Randomly choose when to execute fix fingers
                execute_fix_fingers := get_random_int() % 4 == 0
                execute_stabilize_ring := get_random_int() % 3 == 0
                //Perform join-ring action
                if message.Do == "join-ring" {

					//log.Printf("\n***** JOIN-RING INSIDE NET_NODE *****\n")
                    
                    if val, ok := ring_nodes.Load(channel_id); ok != true {
                        _ = val
                        sponsoring_node_id := message.SponsoringNode
                        Join_ring(sponsoring_node_id, &node_obj)
                        ring_nodes.Store(channel_id, &node_obj)
                    }else{
                        log.Printf("\nNode %d is already in the ring; cannot join-ring\n", channel_id)
                    }

                    if test_mode == true {
                        test_channel <- "Done"
                    }

                   } else if message.Do == "leave-ring" {
                    if val, ok := ring_nodes.Load(channel_id); ok{
                        _ = val
                        Leave_ring(&node_obj, message.Mode)
                        ring_nodes.Delete(channel_id)
                    }else{
                        log.Printf("\nNode %d is not in the ring; cannot leave-ring\n", channel_id)
                    }
                    if test_mode == true {
                        test_channel <- "Done"
                    }
                //tells the node_obj that the respond to may be its predecessor
                } else if message.Do == "ring-notify" {
                    Notify(&node_obj, message.RespondTo)

                } else if message.Do == "find-ring-successor" {
                    //respond-to contains the "sponsor" of this request
                    //respond-to is the node that recieves the answer of find ring successor
                    if sponsor_node, ok := ring_nodes.Load(message.RespondTo); ok{
                        FindRingSuccessor(sponsor_node, message.TargetId, message.RespondTo)
                    }else{
                        log.Printf("\nRespondTo node: %d is not responding...not in ring?\n", message.RespondTo)
                    }
                } else if message.Do == "find-ring-predecessor" {
                    //Tell node_obj to find the predecessor of target id and report back to respond-to
                    FindRingPredecessor(&node_obj, message.TargetId, message.RespondTo)
                }else if message.Do == "fix-ring-fingers"{
                    FixRingFingers(&node_obj)
                    if test_mode == true {
                        test_channel <- "Done"
                    }
                //The node that recieves this message is the node
                //That needs to have its fingers built.
                //the respond-to field added on is the nodes successor
                }else if message.Do == "init-ring-fingers"{
                    //InitRingFingers(&node_obj, message.RespondTo)
                    
                }else if message.Do == "get" {
                    GetData(&node_obj, node_obj.ChannelId, message.Data.Key)
                    if test_mode == true {
                        test_channel <- "Done"
                    }
                }else if message.Do == "remove" {
                    RemoveData(&node_obj, node_obj.ChannelId, message.Data.Key)
                    if test_mode == true {
                        test_channel <- "Done"
                    }
                }else if message.Do == "put" {
                    PutData(&node_obj, node_obj.ChannelId, message.Data.Key, message.Data.Value)
                    if test_mode == true {
                        test_channel <- "Done"
                    }
                }else if message.Do == "stabilize-ring"{
                    Stabilize(&node_obj)
                    if test_mode == true {
                        test_channel <- "Done"
                    }            

                }else if message.Do == "find-closest-preceeding-node" {
                    //Have node_obj find the closest preceeding node to the target_id
                    FindClosestPreceedingNode(&node_obj, message.TargetId)

                //Tell node to set its successor to target-id
                //{do: set-successor, target-id: target-id}
                }else if message.Do == "set-successor" {
                    //Set the successor as the target id
                    node_obj.Successor = message.TargetId

                //Give the predecessor to the respond-to node
                // {do: get-predecessor, respond-to: respond-to}
                }else if message.Do == "get-predecessor" {
                    var bucket_msg =  msg.BucketMessage {Identifier: node_obj.Predecessor}
                    string_message, err := json.Marshal(bucket_msg)
                    check_error(err)
                    SendDataToBucket(message.RespondTo, string(string_message))
                }else if message.Do == "check-predecessor"{
                    CheckPredecessor(&node_obj)
                }

                //Randomly cause a to fix fingers/execute stabilize
                //Only do this randomly in non-test mode
                random_ring_node := get_random_ring_node()
                if random_ring_node != -1 && test_mode == false{
                    if execute_fix_fingers == true {
                        SendDataToNetwork(random_ring_node, "{\"do\": \"fix-ring-fingers\"}")
                    }else{

                        if execute_stabilize_ring == true {
                            //SendDataToNetwork(random_ring_node, "{\"do\": \"check-predecessor\"}")
                            SendDataToNetwork(random_ring_node, "{\"do\": \"stabilize-ring\"}")
                        }
                    }
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
                time.Sleep(5)
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

    var test_m = prog_args[0]
    var file_name = prog_args[1]
    mean_wait, err := strconv.ParseFloat(prog_args[2], 64)
    check_error(err)
    var num_nodes int

    //Set up test mode
    if test_m == "YES"{
        test_mode = true
    }else{
        test_mode = false
        num_nodes, _ = strconv.Atoi(prog_args[3])
        number_of_network_nodes = int(math.Exp2(float64(num_nodes)))
    }

    mean_wait_value = mean_wait
    //The number of nodes is 2^N where N is what the user entered for the amount of nodes
    log.Printf("\nStarting program with the following:\nInstruction File: %s\nNumber of network nodes: %d\nMean wait time: %f\n", file_name, number_of_network_nodes, mean_wait_value)
    log.Println("This is the coordinator.")
    //Create a bunch of random nodes for the network
    init_topology()

    //Get a random ring nodes id
    var random_ring_id int64

    //get a list of string json instructions to send to random nodes
    var instructions []string = create_message_list(file_name)


	//log.Printf("\nRaw instructions: %s\n", instructions)
    var channel_id int64
    for i := 0; i < len(instructions); i++ {
        //pick a random node in the ring to send the message to.
        prev_random_ring_id := int64(-1)
        random_ring_id = get_random_ring_node()
        random_network_id := get_random_network_node()
            byte_msg := []byte(instructions[i])

			//log.Printf("\n****Byte msg Instruction: %s\n", instructions[i])
            var message msg.Message

            err := json.Unmarshal(byte_msg, &message)
		    //check_error(err)	

			//log.Printf("\n@@@@@@------ UnMarshal Struct Message: %+v -------@@@@@@@\n", message)

            if err != nil {
                log.Println("Reached the end of the json instructions")
                break
            }

			//log.Printf("\n@@@@@@@---- Message.Do: %s @@@@@@-----\n", message.Do) 

            //format join ring instruction with random sponsoring node
            if message.Do == "join-ring" {
                random_ring_id = -1
                for random_ring_id == -1 {
                    random_ring_id = get_random_ring_node()
                }

                if test_mode == false {
                    message.SponsoringNode = random_ring_id
                }
                prev_random_ring_id = random_ring_id
                _ = prev_random_ring_id

                if test_mode == false {
                    channel_id = random_network_id
                }else {
                    channel_id = message.TestSendTo
                }

            }else if message.Do == "fix-ring-fingers" {

                //check test mode
                if test_mode == false {
                    channel_id = random_ring_id
                }else{
                    channel_id = message.TestSendTo
                }
                if channel_id < 0 {
                    log.Println("There is no node in the ring to fix fingers")
                    continue
                }
            }else{

                //check test mode
                if test_mode == false {
                    channel_id = random_ring_id
                }else{
                    channel_id = message.TestSendTo
                }

            }

            modified_inst, err := json.Marshal(message)
            check_error(err)
            // Give a random node instructions
            map_lock.Lock()

            //Send the instruction
            network[channel_id] <- string(modified_inst)
            map_lock.Unlock()
            //if test mode, wait until instruction is done before sending another
            if test_mode == true{
                _ = <-test_channel
            }else{
                time.Sleep(3)
            }
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
        //Test mode... the default number of nodes in the ring is 2^2 and the default first node is 2
        //sponsoring node in the file must be filled
        log.Println("USAGE1 Test Mode: go run main.go <TEST_MODE> <INST FILE> <AVG_WEIGHT_TIME>")
        log.Println("USAGE2 Non-Test Mode: go run main.go  <TEST_MODE> <INST FILE> <AVG_WEIGHT_TIME> <N WHERE #NODES is 2^N>")
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



