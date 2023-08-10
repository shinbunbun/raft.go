package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/rpc"

	"github.com/secamp-y3/raft.go/domain"
)

const PORT = "8080"

func infoHandler(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Query()
	address := param.Get("address")

	// fmt.Printf("address: %s\n", address)

	client, err := rpc.Dial("tcp", address)
	if err != nil {
		fmt.Println("dialing: ", err)
	}
	input := domain.GetArgs{}
	var reply domain.GetReply
	// fmt.Printf("input: %+v\n", input)
	err = client.Call("StateMachine.Get", input, &reply)
	// fmt.Printf("reply: %+v\n", reply)
	response, err := json.Marshal(reply)
	if err != nil {
		fmt.Println("marshaling:", err)
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,UPDATE,OPTIONS")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(response))
}

func monitorHandler(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Query()
	address := param.Get("address")

	fmt.Printf("address: %s\n", address)

	client, err := rpc.Dial("tcp", address)
	if err != nil {
		fmt.Println("dialing: ", err)
	}
	input := domain.FetchStateArgs{}
	var reply domain.FetchStateReply
	fmt.Printf("input: %+v\n", input)
	err = client.Call("Monitor.FetchState", input, &reply)
	fmt.Printf("reply: %+v\n", reply)
	response, err := json.Marshal(reply)
	if err != nil {
		fmt.Println("marshaling:", err)
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,UPDATE,OPTIONS")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(response))
}

func appendHandler(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Query()
	address := param.Get("address")
	log := param.Get("log")

	fmt.Printf("address: %s\n", address)

	client, err := rpc.Dial("tcp", address)
	if err != nil {
		fmt.Println("dialing: ", err)
	}
	input := domain.AppendLogsArgs{Entries: []domain.Log{{Log: log}}}
	var reply domain.AppendLogsReply
	fmt.Printf("input: %+v\n", input)
	err = client.Call("StateMachine.AppendLogs", input, &reply)
	fmt.Printf("reply: %+v\n", reply)
	response, err := json.Marshal(reply)
	if err != nil {
		fmt.Println("marshaling:", err)
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,UPDATE,OPTIONS")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(response))
}

func main() {
	http.HandleFunc("/info", infoHandler)
	http.HandleFunc("/monitor", monitorHandler)
	http.HandleFunc("/append", appendHandler)
	err := http.ListenAndServe("0.0.0.0:"+PORT, nil)
	if err != nil {
		panic(err)
	}
}
