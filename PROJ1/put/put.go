package put

import msg "./utils/message_defs"
import node "./utils/node_defs"
//import "bytes"


//{"do": "put", "data": { "key" : "a key", "value" : "a value" }, "respond-to": "channel-id"}
// instructing the receipient node to store the given (key,value) pair in the appropriate ring node. 

// func FindClosestPreceedingNode(node_obj *node.Node, target_id int64)

/*

func GetDataFromBucket(node_id int64)(bucket_data string){
func ExtractIdFromBucketData(data string)(identifier int64){

*/

func Put(data *msg.Data, respond_to int64, node_obj *node.Node) {

		// FindPReceedingNode
		var closest_node = FindClosestPreceedingNode(&node_obj.FingerTable, data.Key)

		// Get ID from bucket
		var closest_id = ExtractIdFromBucketData(data.Key)

		// Get data from bucket
		var closest_data = GetDataFromBucket(closest_id)

		// Put data in node
		closest_node.DataTable[data.Key] := closest_data
	}
}
