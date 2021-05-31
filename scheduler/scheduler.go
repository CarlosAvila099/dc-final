package scheduler

import (
	"context"
	"log"
	"time"

	pb "github.com/CarlosAvila099/dc-final/proto"
	"github.com/CarlosAvila099/dc-final/resources"
	"google.golang.org/grpc"
)

//const (
//	address     = "localhost:50051"
//	defaultName = "world"
//)


func schedule(job resources.Job) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(job.Address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	log.Println(job.RPCName)

	fr, err := c.Filter(ctx, &pb.FilterRequest{Filter: job.RPCName, Image: job.ImagePath, Workload: job.WorkloadPath, Counter: job.CurrentId})
	if err != nil {
		log.Fatalf("could not filter: %v", err)
	}
	log.Printf("Scheduler: RPC respose from %s : %s", job.Address, fr.GetMessage())
}

func Start(jobs chan resources.Job) error {
	for {
		job := <-jobs
		schedule(job)
	}
	return nil
}
