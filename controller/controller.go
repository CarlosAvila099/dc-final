package controller

import (
	"fmt"
	"strconv"
	"strings"
	"encoding/json"
	"github.com/CarlosAvila099/dc-final/resources"

	// register transports
	_ "go.nanomsg.org/mangos/transport/all"
)

func createWorkload(workloadName string, workloadFilter string) (int, bool) {
	var filterCode int
	if workloadFilter == ""{
		filterCode = resources.NORMAL
	} else if workloadFilter == "grayscale"{
		filterCode = resources.GRAYSCALE
	} else if workloadFilter == "blur"{
		filterCode = resources.BLUR
	} else {
		return -1, false
	}
	id := len(workloadManager)
	wl := resources.Workload{id, filterCode, workloadName, resources.SCHEDULING, 0, make([]int, 0)}
	workloadManager = append(workloadManager, wl)
	return id, true
}

func getWorkloads() string{
	workloadString := "["
	if len(workloadManager) < 1{
		workloadString = "There are no active workloads"
		return workloadString
	}
	for _, workload := range workloadManager{
		workloadString += strconv.Itoa(workload.Id) + ", "
	}
	workloadString = workloadString[:len(workloadString) - 2] + "]"
	return workloadString
}

func getMeaning(msg string) (int, []string){
	splitted := strings.Split(msg, "&")
	operation, _ := strconv.Atoi(splitted[0])
	restMsg := splitted[1:]
	return operation, restMsg
}

func operate(o int, msg []string) (resources.Workload, string) {
	var ok bool
	var id int
	var work = resources.Workload{}
	var response = ""
	switch(o){
		case 1:
			name := msg[0]
			filter := msg[1]
			if id, ok = createWorkload(name, filter); !ok{
				response = "-1"
				break
			}
			work =  workloadManager[id]
			break
		case 2:
			id, _ := strconv.Atoi(msg[0])
			if len(workloadManager) <= id{
				response =  "-2"
				break
			}
			work = workloadManager[id]
			break
		case 3:
			response = getWorkloads()
			break
	}
	return work, response
}

var workloadManager []resources.Workload //Manages workload id
//var imageManager []int //Manages workload id

func Start() {
	var operation int
	var splitted []string
	var cr = resources.ControllerResponse {}
	socket := resources.GetSocket(true)
	for {
		msg := resources.ReceiveFromPair(socket)
		fmt.Println(msg)
		operation, splitted = getMeaning(msg)
		work, response := operate(operation, splitted)
		if response == ""{
			cr = resources.ControllerResponse{ work, "" }
		} else {
			cr = resources.ControllerResponse{ resources.Workload{}, response }
		}
		resp, _ := json.Marshal(cr)
		resources.SendToPair(socket, string(resp))
	}
}
