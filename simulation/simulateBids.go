package main

import (
	"context"
	"log"
	"sync"
	"time"

	proto "Replication/grpc" // Import your generated proto package

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to the auction server
	serverAddress := ":50051" // Update this if using a different server port
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := proto.NewAuctionServerClient(conn)

	// Use WaitGroup to synchronize goroutines
	var wg sync.WaitGroup

	// Define the users and their bids
	users := []struct {
		name   string
		amount int32
	}{
		{"Alice", 100},
		{"Bob", 150},
	}

	wg.Add(len(users)) // Add the number of concurrent users

	// Simulate each user bidding concurrently
	for _, user := range users {
		go func(userName string, bidAmount int32) {
			defer wg.Done()
			bid(client, userName, bidAmount)
		}(user.name, user.amount)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Fetch the auction result after all bids
	result, err := client.Result(context.Background(), &proto.Empty{})
	if err != nil {
		log.Fatalf("Failed to fetch auction result: %v", err)
	}
	log.Printf("Auction result: %s, Highest Bid: %d", result.Result, result.HighestBid)
}

func bid(client proto.AuctionServerClient, bidder string, amount int32) {
	// Send a bid
	req := &proto.Amount{
		Amount:    amount,
		Bidder:    bidder,
		Timestamp: int32(time.Now().UnixNano()), // Add a timestamp for ordering
	}
	resp, err := client.Bid(context.Background(), req)
	if err != nil {
		log.Printf("Error while bidding for %s: %v", bidder, err)
		return
	}
	log.Printf("Response for %s: %s", bidder, resp.Ack)
}
