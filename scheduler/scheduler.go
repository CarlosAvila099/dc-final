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
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: job.RPCName})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Scheduler: RPC respose from %s : %s", job.Address, r.GetMessage())

	r, err = c.GrayscaleFilter(ctx, &pb.HelloRequest{Name: job.RPCName})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Scheduler: RPC respose from %s : %s", job.Address, r.GetMessage())
}

func Start(jobs chan resources.Job) error {
	for {
		job := <-jobs
		schedule(job)
	}
	return nil
}
