package main

import (
	"context"
	"log"
	"net"
	"os"
	"sync"
	"time"

	proto "Replication/grpc"

	"google.golang.org/grpc"
)

const auctionDuration = 100 * time.Second

type AuctionServer struct {
	proto.UnimplementedAuctionServerServer
	highestBid    int
	highestBidder string
	bidders       map[string]bool
	isAuctionOver bool
	mutex         sync.Mutex
}

func main() {
	// to the log
	file, err := os.OpenFile("auction_log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer file.Close()

	log.SetOutput(file)

	// actual main
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	auctionServer := &AuctionServer{
		highestBid:    0,
		bidders:       make(map[string]bool),
		isAuctionOver: false,
	}
	proto.RegisterAuctionServerServer(grpcServer, auctionServer)

	auctionServer.AuctionTimer()

	log.Printf("Server is running at %v", listener.Addr())
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
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
		// s.highestBidder = whatever highestbidder ID is
		return &proto.Ack{
			Ack: "success",
		}, nil
	} else {
		return &proto.Ack{
			Ack: "BidException: yOu PoOR bAsTArD",
		}, nil
	}
}

func (s *AuctionServer) Result(ctx context.Context, req *proto.Empty) (*proto.Outcome, error) {
	s.mutex.Lock()
	defer s.mutex.Lock()

	if s.isAuctionOver {
		return &proto.Outcome{
			Result:     "Auction over, the highest bidder was " + s.highestBidder,
			HighestBid: int32(s.highestBid),
		}, nil
	} else {
		return &proto.Outcome{
			Result:     "Auction is ongoing, the highest bidder is " + s.highestBidder,
			HighestBid: int32(s.highestBid),
		}, nil
	}
}

func (s *AuctionServer) AuctionTimer() {
	time.Sleep(auctionDuration) //makes the auction run for an amount of time
	s.mutex.Lock()
	s.isAuctionOver = true //ends auction
	s.mutex.Unlock()
	log.Println("Auction has ended")
}
