package message_defs


/*
This struct defines the json messages for the
CHORD protocol.

Fields:
do - The action for the node to take
sponsoring_node - The sponsoring node already in the coord ring for the action.
mode - The mode for leaving the CHORD ring
respond_to - The node (represented as a channel id) to direct the given action towards
data - The key/value pair in the ring representing a hash entry.
test-send-to - This field is only used for testing..this is the node id to send the message to
*/
type Message struct {
	Do string `json:"do"`
	SponsoringNode int64 `json:"sponsoring-node"`
	Mode string `json:"mode"`
	RespondTo int64 `json:"respond-to"`
	TargetId int64 `json:target-id`
	Data Data `json:"data"`
	/*Data struct {
		Key string `json:"key"`
		Value string `json:"value"`
	}*/
	TestSendTo int64 `json:"test-send-to"`
}

type Data struct {
	Key string `json:"key"`
	Value string `json:"value"`
}

/*
The tag is a field that nodes use to ensure that the order
of the messages are correct
*/
type BucketMessage struct {
	Identifier int64 `json:"identifier"`
}
