package main

import (
	"log"
	"time"

	"github.com/CarlosAvila099/dc-final/api"
	"github.com/CarlosAvila099/dc-final/controller"
	"github.com/CarlosAvila099/dc-final/scheduler"
	"github.com/CarlosAvila099/dc-final/resources"
)

func main() {
	log.Println("Welcome to the Distributed and Parallel Image Processing System")
	jobs := make(chan resources.Job)

	// Start Controller
	go controller.Start(jobs)

	// Start Scheduler
	go scheduler.Start(jobs)

	time.Sleep(time.Second)
	// API
	go api.Start()
	for{
		
	}
}
