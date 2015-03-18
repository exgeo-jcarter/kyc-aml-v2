package main

import (

)

func main() {
	
	server, err := NewKycAmlServer("config.json")
	if err != nil {
		return
	}
	
	server.FuzzyTrain()
	
	err = server.Listen()
	if err != nil {
		return
	}
}

/*
func SocketMgr() {
	
	listener, _ := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	defer listener.Close()
	
	for {
		
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	
	for {
		
		connbuf := bufio.NewReader(conn)
		
		res, err := connbuf.ReadBytes('\n')
		if err != nil {
			continue
		}
		
		var socketMsg SocketMsgS
		err = json.Unmarshal(res[:len(res)-1], &socketMsg)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		
		switch socketMsg.Action {
			
		case "load_dataset":
			res, err = LoadDataset()
			if err != nil {
				continue
			}
			
		case "query":
			res, err = Query(socketMsg.Value[0], socketMsg.Value[1], socketMsg.Value[2])
			if err != nil {
				continue
			}
		}
		
		conn.Write([]byte(string(res) + "\n"))
	}
	
	conn.Close()
}


func main() {

	LoadDataset()
	SocketMgr()
}
*/