package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	proto "Replication/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuctionServer struct {
	proto.UnimplementedAuctionServerServer
	highestBid    int
	highestBidder string
	bidders       map[string]bool
	isAuctionOver bool
	mutex         sync.Mutex
	reps          []string
	port          string
}

const auctionDuration = 1000 * time.Second

var port = flag.String("port", "50051", "Server port")

func main() {
	// do it for the log
	file, err := os.OpenFile("../auction_log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer file.Close()

	log.SetOutput(file)

	// handle 'crashing' for the log
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	flag.Parse()

	// actual main
	log.Println("i want to start listening")
	listener, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("Listener created successfully: %v", listener.Addr())

	grpcServer := grpc.NewServer()
	auctionServer := &AuctionServer{
		highestBid:    0,
		bidders:       make(map[string]bool),
		isAuctionOver: false,
		port:          *port,
	}
	proto.RegisterAuctionServerServer(grpcServer, auctionServer)

	go func() {
		log.Println("Starting auction timer...")
		auctionServer.AuctionTimer()
		log.Println("Auction timer completed")
	}()

	go func() {
		log.Printf("Server is running at %s", listener.Addr())
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %s", err)
		}
	}()

	// logging the crash/interruption
	<-stop
	log.Printf("System interrupted. Shutting down server %v. \n", listener.Addr())
	grpcServer.GracefulStop()
	log.Printf("Server %v stopped... \n", listener.Addr())
}

func (s *AuctionServer) Bid(ctx context.Context, req *proto.Amount) (*proto.Ack, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isAuctionOver {
		return &proto.Ack{
			Ack: "fail",
		}, nil
	}

	if int(req.Amount) > s.highestBid {

		s.highestBid = int(req.Amount)
		s.highestBidder = req.Bidder

		// propagate the bid request to all replicas
		s.PropagateBid(req)

		return &proto.Ack{
			Ack: "success",
		}, nil
	}

	return &proto.Ack{
		Ack: "BidException: yOu PoOR bAsTArD",
	}, nil

}

func (s *AuctionServer) PropagateBid(req *proto.Amount) {
	for _, addr := range s.reps {
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Cannot connect to replica %v: %v", addr, err)
			continue
		}
		defer conn.Close()

		client := proto.NewAuctionServerClient(conn)
		_, err = client.Bid(context.Background(), req)
		if err != nil {
			log.Printf("Failed to propagate bid to replica %v: %v", addr, err)
		}
		conn.Close()
	}
}

func (s *AuctionServer) Result(ctx context.Context, req *proto.Empty) (*proto.Outcome, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isAuctionOver {
		return &proto.Outcome{
			Result:     "Auction over, the highest bidder was " + s.highestBidder,
			HighestBid: int32(s.highestBid),
		}, nil
	}

	return &proto.Outcome{
		Result:     "Auction is ongoing, the highest bidder is " + s.highestBidder,
		HighestBid: int32(s.highestBid),
	}, nil

}

func (s *AuctionServer) AuctionTimer() {
	time.Sleep(auctionDuration) // makes the auction run for an amount of time
	s.mutex.Lock()
	s.isAuctionOver = true // ends auction
	s.mutex.Unlock()
	log.Println("Auction has ended")
}
