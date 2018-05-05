# Projects
Projects
Project 1 Implementing Chord

Authors: Nick Allgood, Mariama Adangwa, and  Mohit Khatwani

TABLE OF CONTENTS
------------------
+ Basic Instructions
+ MESSAGE CHANNELS
  + BUCKET CHANNEL
  + COMMAND (Network) CHANNEL
+ MESSAGE TYPES
  + BUCKET MESSAGE TYPES
  + COMMAND (Network) MESSAGE TYPES
+ TESTING PROJECT 1
  + NON-TEST MODE
  + TEST MODE

Basic Instructions:
NOTE: Occassionally code may deadlock at the very beginning.
Keep running the "go run" command until this does not happen
To run and see debug messages run the program as follows:
Non-Test Mode: go run main.go <INST FILE> <TEST_MODE> <AVG_WAIT_TIME> <N WHERE #NODES is 2^N>
Test Mode: go run main.go <INST FILE> <TEST_MODE> <AVG_WAIT_TIME>

+++MESSAGE CHANNELS+++
----------------------
The 2 types of message channels are:

BUCKET CHANNELS
Used for sending/receiving data between nodes
Two notable functions for bucket channels are:
- SendDataToBucket - Sends the given string json data message to the designated channel id (non-blocking)
- GetDataFromBucket - Listens on the specified bucket channel for incoming data messages (blocking)
  By default, the GetDataFromBucket timeout after 20 seconds of waiting for messages.

COMMAND (Network) CHANNELS
Used for sending/receiving commands between nodes
Some notable functions for network channels are:
- SendDataToNetwork - Sends the given json command message to the specified node that processes messages
                    - on the given channel id.

+++MESSAGE TYPES+++
-------------------
The 2 types of message types are used for the 2 types of channels: Network channels (COMMAND) and bucket channels
(BUCKET).

BUCKET MESSAGE TYPES
The BUCKET is used for sending and receiving data across the network. The BUCKET channel has the following message definitions:
{"identifier": "channel-id"}        - A message in which a channel id is given in the 'identifier' field.

COMMAND (Network) MESSAGE TYPES
The COMMAND (Network) Channel is used for sending and receiving commands.
The COMMAND channel has the following message definitions:

{"do": "join-ring", "sponsoring-node": "channel-id" }
{"do": "leave-ring" "mode": "immediate or orderly"}
{"do": "stabilize-ring" }
{"do": "init-ring-fingers" }
{"do": "fix-ring-fingers" }
{"do": "ring-notify", "respond-to": "channel-id" }
{"do": "get-ring-fingers", "respond-to": "channel-id" }
{"do": "find-ring-successor", "respond-to": "channel-id"}
{"do": "find-ring-predecessor", "respond-to": "channel-id"}
{"do": "put", "data": { "key" : "a key", "value" : "a value" }, "respond-to": "channel-id"}
{"do": "get", "data": { "key" : "a key" }, "respond-to": "channel-id"}
{"do": "remove", "data": { "key" : "a key" }, "respond-to": "channel-id"}

Additional message definitions:
{"do": "set-successor", "target-id":"channel-id" }                - The recipient node that receives 
                                                                  - this message sets its successor to the target-id.
{"do": "find-closest-preceeding-node", "target-id":"channel-id"}  - The recipient node that receives
                                                                  - this message uses its finger table to find the 
                                                                  - closest preceeding node of the target-id.


+++TESTING PROJECT 1+++

NON-TEST MODE
-------------
The json instructions for the coordinator should be in the instructions.txt file.
go run main.go NO instructions.txt 3 2 
//mean timeout before nodes process instructions is 3.
//Runs the program with instructions from instructions.txt, creates 2^2 nodes.
//Instructions can happen in a random order

TEST MODE
---------
The json instructions for the coordinator should be in the test_*.txt file.
go run main.go YES test_instructions.txt 3 
//mean timeout before nodes process instructions is 3.
//Runs the program with instructions from test_instructions.txt, creates 2^2 nodes by default,
//The default can be changed by changing the number_of_network_nodes global variable
//The default first node in test mode is 2 but can be changed by modifying the test_first_node global variable
//Instructions can happen in a non-random order

There are several test_* files in the PROJ1 directory. Each of these files are the test instruction files for the project, and are to be run in test mode:
- test_put.txt
- test_get.txt
- test_leave.txt
- test_join.txt
- test_stabilize.txt
- test_fix_fingers.txt
