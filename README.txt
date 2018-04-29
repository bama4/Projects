# Projects
Projects
Project 1 Implementing Chord

Authors: Nick Allgood, Mariama Adangwa, and  Mohit Khatwani

Instructions:
To run and see debug messages run the program as follows:
Non-Test Mode: go run main.go <INST FILE> <TEST_MODE> <AVG_WAIT_TIME> <N WHERE #NODES is 2^N>
Test Mode: go run main.go <INST FILE> <TEST_MODE> <AVG_WAIT_TIME>


The json instructions for the coordinator should be in the instructions.txt file.
For join-ring, the sponsoring_id is randomly generated
Example:

//Non test mode
go run main.go NO instructions.txt 3 2 
//mean timeout before nodes process instructions is 3.
//Runs the program with instructions from instructions.txt, creates 2^2 nodes, and the

//test mode
go run main.go YES test_instructions.txt 3 
//mean timeout before nodes process instructions is 3.
//Runs the program with instructions from test_instructions.txt, creates 2^2 nodes by default,
//The default can be changed by changing the number_of_network_nodes global variable
//The default first node in test mode is 2 but can be changed by modifying the test_first_node global variable
