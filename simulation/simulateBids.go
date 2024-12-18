package main

import (
	"context"
	"log"
	"sync"
	"time"

	proto "Replication/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	serverAddress := ":50051"
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := proto.NewAuctionServerClient(conn)

	var wg sync.WaitGroup

	users := []struct {
		name   string
		amount int32
	}{
		{"Anna", 100},
		{"Karoline", 100},
	}

	wg.Add(len(users))

	for _, user := range users {
		go func(userName string, bidAmount int32) {
			defer wg.Done()
			bid(client, userName, bidAmount)
		}(user.name, user.amount)
	}

	wg.Wait()

	result, err := client.Result(context.Background(), &proto.Empty{})
	if err != nil {
		log.Fatalf("Failed to fetch auction result: %v", err)
	}
	log.Printf("Auction result: %s, Highest Bid: %d", result.Result, result.HighestBid)
}

func bid(client proto.AuctionServerClient, bidder string, amount int32) {
	req := &proto.Amount{
		Amount:    amount,
		Bidder:    bidder,
		Timestamp: int32(time.Now().UnixNano()),
	}
	resp, err := client.Bid(context.Background(), req)
	if err != nil {
		log.Printf("Error while bidding for %s: %v", bidder, err)
		return
	}
	log.Printf("Response for %s: %s", bidder, resp.Ack)
}
