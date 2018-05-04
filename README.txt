# Projects
Projects
Project 1 Implementing Chord

Authors: Nick Allgood, Mariama Adangwa, and  Mohit Khatwani

TABLE OF CONTENTS
------------------
+ Basic Instructions
+ MESSAGE TYPES
+ TESTING PROJECT 1

Basic Instructions:
To run and see debug messages run the program as follows:
Non-Test Mode: go run main.go <INST FILE> <TEST_MODE> <AVG_WAIT_TIME> <N WHERE #NODES is 2^N>
Test Mode: go run main.go <INST FILE> <TEST_MODE> <AVG_WAIT_TIME>


+++MESSAGE TYPES+++
-------------------
The 2 types of message types are used for the 2 types of channels: Network channels (COMMAND) and bucket channels
(BUCKET).
The BUCKET is used for sending and receiving data across the network. The BUCKET channel has the following message definitions:


The COMMAND (Network) Channel is used for sending and receiving commands. The COMMAND channel has the following message definitions:



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

Some additional message formats that can be seen are as follows:
BUCKET MESSAGES:

NETWORK MESSAGES:
