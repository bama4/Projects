# Projects
Projects
Project 1 Implementing Chord

Authors: Nick Allgood, Mariama Adangwa, and  Mohit Khatwani

Instructions:
To run and see debug messages run the program as follows:
go run main.go instructions.txt <NUM_OF_NODES AS N (GENERATES 2^N Nodes)> <Mean Timeout>

The json instructions for the coordinator should be in the instructions.txt file.
For join-ring, the sponsoring_id is randomly generated
Example:

go run main.go instructions.txt 3 2 
//Runs the program with instructions from instructions.txt, creates 2^3 nodes, and the
//mean timeout before nodes process instructions is 2.
