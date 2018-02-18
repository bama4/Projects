package main

import "encoding/json"
import "os"
import "io"
import "strings"
import "strconv"
import "bytes"
import "sync"
import "log"
import "fmt"
import "time"

var comms_in = make(chan string)
var comms_out = make(chan string)
var comms_exit = make(chan int)
var global_is_logging_file bool
var global_sum = int64(0)
var global_wg sync.WaitGroup
var global_num_threads int

type Request struct {
	Datafile string `json:"datafile"`
	Start int `json:"start"`
	End int `json:"end"`
}

type Response struct {
	Value int64 `json:"value"`
	Prefix string `json:"prefix"`
	Suffix string `json:"suffix"`
}

/*
Calculates the sum of an array
*/
func calc_sum(slice []int64) (sum int64) {
	sum = 0
	for i := 0; i < len(slice); i++ {
		sum += slice[i]
	}
	log.Println("Calculated the sum as ", sum)
	return
}

/*
Processes the sum message. This is the worker thread
*/
func process_sum_message(msg string) () {
	defer global_wg.Done()
	byte_msg := []byte(msg)
	var request Request

	err := json.Unmarshal(byte_msg, &request)
	if err != nil {
		log.Println("Error decoding msg: ", msg)
	}
	file_name := request.Datafile
	start_pos := request.Start
	end_pos := request.End
	// Recieve integers from file
	int_slice := get_ints_from_file(file_name, start_pos, end_pos)
	// Calculate sum of slice	
	sum := calc_sum(int_slice)
	global_sum += sum
	//Send sum response message
	resp_msg := "0"
	if len(int_slice) > 0 {	
		resp_msg = create_sum_response(sum, strconv.FormatInt(int_slice[0], 10),strconv.FormatInt(int_slice[len(int_slice)-1], 10))
	}else {
		resp_msg = create_sum_response(sum, "0", "0")
	}

	comms_out <-resp_msg
}

/*
This is the process that handles the worker responses.
*/
func process_worker_response(resp string){
	log.Println("Got message: ", resp)
	byte_msg := []byte(resp)
	var response Response

	err := json.Unmarshal(byte_msg, &resp)
	if err != nil {
		log.Println("Error decoding msg: ", resp)
	}
	global_sum += response.Value
	log.Println("Global Sum updated to: ", global_sum)
	if global_is_logging_file == true {
		fmt.Println("Global Sum updated to: ", global_sum)
	}
}

/*
Create the json response message.
*/
func create_sum_response(val int64, pre string, suf string) (resp string){
	var response = &Response {Value: val, Prefix: pre, Suffix: suf}
	resp_byte, err := json.Marshal(response)
	if err != nil {
		log.Println("Error decoding msg: ", err)
	}
	resp = string(resp_byte)
	return
}

/*
Given the file descriptor and the number of slices, this function
parses the file. If it hits the middle of a number at the start position, it ignores that number.
If it hits the middle of the number at the end position, then it takes the entire number.
*/
func get_ints_from_file(file_name string, start int, end int) []int64{
	var int_arr []int64
	file, err := os.Open(file_name)
	defer file.Close()
	if err != nil {
		log.Println("Could not open file: ", file_name)
		log.Println(err)
		os.Exit(1)
	}
	
	//Get file size
	file_stat, err_stats := os.Stat(file_name)
	if err_stats != nil {
		log.Println("Could not get file stats: ", err_stats)
	}
	file_size := file_stat.Size()

	var new_line int = 0
	var space int = 1
	delimit_chars := []byte{'\n', ' '}
	
	byte_buffer := make([]byte, 1)
	
	var file_ctr int = start
	if start != 0{
		//check prefix
		log.Println("Checking prefix...")
		for byte_buffer[0] != delimit_chars[new_line] && byte_buffer[0] != delimit_chars[space] {
			log.Println("Current byte: ", byte_buffer[0])

			offset, err := file.Seek(int64(file_ctr), io.SeekStart)
			log.Println("Offset from file seek ", offset)
			if err != nil {
				log.Println("Error seeking from file: %s", err)
				break
			}

			num_read, err := file.Read(byte_buffer)
			if err != nil {
				log.Println("Error reading %d bytes from file: %s", num_read, err)
				break
			}
			file_ctr += 1
		}

		//Update start position
		start = file_ctr	
	}


	//Reset seek pointer to the beginning of the file
	offset, err := file.Seek(0, io.SeekStart)
	log.Println("Offset from file seek ", offset)
	if err != nil {
		log.Println("Error seeking to the beginning of the file: %s", err)
	}


	byte_buffer[0] = byte(0)

	file_ctr = end + 1
	//check suffix
	if int64(end) != file_size {
		log.Println("Checking suffix...")
		for byte_buffer[0] != delimit_chars[new_line] && byte_buffer[0] != delimit_chars[space] {
			//check the next character
			log.Println("Current byte: ", byte_buffer[0])			
			offset, err := file.Seek(int64(file_ctr), io.SeekStart)
			log.Println("Offset from file seek", offset)
			if err != nil {
				log.Println("Error seeking from file: %s", err)
				break
			}

			num_read, err := file.Read(byte_buffer)
			if err != nil {
				log.Println("Error reading %d bytes from file: %s", num_read, err)
				break
			}
			file_ctr += 1
		}
		end = file_ctr
	}

	bytes_read := end - start
	if bytes_read < 0 {
		return int_arr
	}
	log.Println("Bytes Read: ", bytes_read)
	buffer := make([]byte, bytes_read)

	//Go to starting position
	offset_start, err_start := file.Seek(int64(start), io.SeekStart)
	if err_start != nil {
		log.Println("Error seeking from file: %s", err_start)
	}
	log.Println("Offset: ", offset_start)
	log.Println("New start: ", start)
	log.Println("New end: ", end)
	if err != nil {
		log.Println("Error seeking from file: %s", err)
		return nil
	}

	//Read file in from start to end
	num_read, err := file.Read(buffer)
	if err != nil {
		log.Println("Error reading %d bytes from file: %s", bytes_read, err)
		return nil
	}

	file.Close()
	log.Println("Number of bytes read: ", num_read)
	string_buffer := string(buffer)
	string_buffer = strings.Replace(string_buffer, "\n", " ", -1)
	string_slice := strings.Split(string_buffer, " ")
	log.Println("String array from file: ", string_slice)
	for i := 0; i < len(string_slice); i++ {
		str_int64, err := strconv.ParseInt(strings.TrimSpace(string_slice[i]), 10, 64)
		if err != nil {
			log.Println("Error converting string to int: ", err)
			continue
		}
		int_arr = append(int_arr, str_int64)
	}
	log.Println("Created the following int64 array: ", int_arr)
	return int_arr
}

/*
Selects on recieving
*/
func select_on_channel() {
	defer global_wg.Done()
	for true{
		select {
			case msg_recv := <-comms_in:
				//process messages coming in from
				//coordinator
				log.Println("Message recieved from comms_in", msg_recv)

				if global_is_logging_file == true {
					fmt.Println("Message recieved from comms_in", msg_recv)
				}
				global_wg.Add(1)
				go process_sum_message(msg_recv)

			case msg_sent := <-comms_out:
				//process response from worker threads
				log.Println("Message sent to comms_out", msg_sent)

				if global_is_logging_file == true {
					fmt.Println("Message sent to comms_out", msg_sent)
				}
				process_worker_response(msg_sent)
				global_num_threads -= 1
				
			default:
				if global_num_threads == 0 {
					return
				}		
		}
	}	
}


/*
Spawns the coordinator thread
*/
func coordinator(prog_args [] string) {
	var num_slices = prog_args[0]
	var file_name = prog_args[1]

	log.Println(num_slices)

	a_file, err := os.Stat(file_name)
	if err != nil {
		log.Println("Could not get file stats: ", err)
		os.Exit(1)
	}

	file_size := a_file.Size()
	num_int_slices, err:= strconv.ParseInt(num_slices, 10, 64)
	global_num_threads = int(num_int_slices)
	if err != nil {
		log.Println("Could not convert number of slices to an int64: ", err)
		os.Exit(1)
	}

	if num_int_slices < 1 {
		log.Println("The NUM_SLICES entered must be greater than 0")
		os.Exit(1)
	}
	
	slice_size := int64(float64(file_size) / float64(num_int_slices) + 0.5)
	log.Println("file size: ", file_size)
	log.Println("slice_size: ", slice_size)
	if num_int_slices > file_size {
		log.Println("Error: Number of slices cannot be greater than the file size!")
		os.Exit(1)
		
	}
	offset := slice_size
	global_wg.Add(1)
	
	//Start coordinator thread (channel watchdog)
	go select_on_channel()
	
	prev_end := int64(0)
	for offset <= file_size {
		var msg bytes.Buffer
		msg.WriteString("{\"datafile\": \"")
		msg.WriteString(file_name)
		msg.WriteString("\", \"start\" : ")
		if prev_end == int64(0) {
			msg.WriteString(strconv.FormatInt(0, 10))
		}else{
			msg.WriteString(strconv.FormatInt(prev_end + int64(1), 10))
		}
		msg.WriteString(", \"end\": ")
		msg.WriteString(strconv.FormatInt(offset, 10))
		prev_end = offset
		offset += slice_size		
		msg.WriteString("}")
		comms_in <- msg.String()
	}
		global_wg.Wait()
		log.Println("Total Sum: ", global_sum)
	}
	
func main(){
	var prog_args = os.Args[1:]
	if len(prog_args) < 2 {
		log.Println("USAGE: go run hw1tn84014.go <NUM_SLICES> <INTEGER_FILE> <DEBUG_FILE (OPTIONAL)>")
		os.Exit(1)
	}
	
	if len(prog_args) == 3 {
		global_is_logging_file = true
		a_file, err := os.OpenFile(prog_args[2], os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Println("Could not open file for logging: ", err)
			os.Exit(1)
		}
		defer a_file.Close()
		log.SetOutput(a_file)
	} else {
		global_is_logging_file = false
	}

	start := time.Now()
	//start coordinator
	coordinator(prog_args)
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Println("Elapsed time: ", elapsed)
	close(comms_in)
	close(comms_out)
	close(comms_exit)
}
