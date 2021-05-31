package controller

import (
	"strconv"
	"strings"
	"encoding/json"

	"github.com/CarlosAvila099/dc-final/resources"

	// register transports
	_ "go.nanomsg.org/mangos/transport/all"
)

func createWorkload(workloadName string, workloadFilter string) (int, bool) {
	var filterCode int
	id := -1
	if workloadFilter == "grayscale"{
		filterCode = resources.GRAYSCALE
	} else if workloadFilter == "blur"{
		filterCode = resources.BLUR
	} else {
		return -1, false
	}
	if workloadName == ""{
		return -2, false
	}
	if !resources.Exists(workloadManager, workloadName){
		id = len(workloadManager)
		wl := resources.Workload{id, filterCode, workloadName, resources.SCHEDULING, 0, make([]int, 0)}
		workloadManager = append(workloadManager, wl)
		ok := wl.SaveWorkload()
		if !ok{
			return -3, false
		}
		return id, true
	}
	return -3, false
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
				if id == -1{
					response = "-1"
					break
				} else if id == -2{
					response = "-1.1"
					break
				} else if id == -3{
					response = "-1.2"
					break
				}
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
		case 4:
			id, _ := strconv.Atoi(msg[0])
			image := msg[1]
			imageType := msg[2]
			if len(workloadManager) <= id{
				response =  "-4.1"
				break
			}
			work = workloadManager[id]
			if imageType == "original"{
				if id, ok = work.CopyImage(image, imageCounter); !ok{
					if id == -1{
						response = "-4.3"
						break
					} else if id == -2{
						response = "-4.4"
						break
					}
				}
				response = strconv.Itoa(id)
			} else if imageType == "filtered"{
				//Creates Job
			} else{
				response =  "-4.2"
				break
			}
		case 5:
			var path string
			var name string
			id, _= strconv.Atoi(msg[0])
			if path, name, ok = resources.SearchImage(workloadManager, strconv.Itoa(id)); !ok{
				response = "-5.1"
				break
			}
			if ok = resources.DownloadImage(path, name); !ok{
				response = "-5.2"
				break
			}
			response = name
	}
	return work, response
}
var (
	workloadManager, imageCounter = resources.ReadWorkloads() //Manages workload id
)

func Start(jobs chan resources.Job) {
	var operation int
	var splitted []string
	var cr = resources.ControllerResponse {}
	socket := resources.GetSocket(true)
	sock := resources.GetPublisher()
	for {
		msg := resources.ReceiveFromPair(socket)
		operation, splitted = getMeaning(msg)
		work, response := operate(operation, splitted)
		if response == ""{
			cr = resources.ControllerResponse{ work, "" }
		} else {
			cr = resources.ControllerResponse{ resources.Workload{}, response }
		}
		resp, _ := json.Marshal(cr)
		resources.SendToPair(socket, string(resp))
		if err := sock.Send([]byte("Hello there")); err != nil {
			resources.Die("Failed publishing: %s", err.Error())
		}
	}
}
