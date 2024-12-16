package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	proto "Replication/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var bidder string

func main() {
	// List of server addresses
	servers := []string{":50051", ":50052", ":50053"}
	var clients []proto.AuctionServerClient

	// Establish connections to all servers
	for _, server := range servers {
		conn, err := grpc.Dial(server, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Failed to connect to server %s: %v", server, err)
			continue
		}
		clients = append(clients, proto.NewAuctionServerClient(conn))
	}

	if len(clients) == 0 {
		log.Fatal("No servers available to connect.")
	}

	log.Println("Connected to servers. Bidding started!")
	input := bufio.NewScanner(os.Stdin)
	log.Println("Enter your username:")
	input.Scan()
	bidder = input.Text()

	// Main loop
	for {
		log.Println("Please write bid [amount] to bid that amount or result :D")
		input.Scan()
		command := strings.TrimSpace(input.Text())
		parts := strings.Split(command, " ")

		if parts[0] == "bid" && len(parts) == 2 {
			amount, err := strconv.Atoi(parts[1])
			if err != nil {
				log.Println("Invalid bid. Usage: bid [amount]")
				continue
			}
			sendBid(int32(amount), clients)
		} else if parts[0] == "result" {
			outcome, err := getResults(clients)
			if err != nil {
				log.Println("Error fetching results:", err)
				continue
			}
			if outcome.Result == "Auction over :O" {
				log.Println("The auction is over!")
				log.Printf("The winner is: %s with a bid of %d\n", outcome.HighestBidder, outcome.HighestBid)
			} else {
				log.Println("The auction is ongoing XD")
				log.Printf("The current highest bid is %d by %s\n", outcome.HighestBid, outcome.HighestBidder)
			}
		} else {
			log.Println("Unknown command, please type bid [amount] or results :/")
		}
	}
}

// Sends a bid to all servers
func sendBid(amount int32, clients []proto.AuctionServerClient) {
	for _, client := range clients {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		req := &proto.Amount{
			Amount:    amount,
			Bidder:    bidder,
			Timestamp: int32(time.Now().UnixNano()),
		}

		ack, err := client.Bid(ctx, req)
		if err != nil {
			log.Println("Failed to send bid to a server:", err)
			continue
		}
		if ack.Ack == "success" {
			log.Println("Bid was successful ;)")
		} else {
			log.Println("Bid failed:", ack.Ack)
		}
	}
}

// Fetches results from all servers and returns the first valid result
func getResults(clients []proto.AuctionServerClient) (*proto.Outcome, error) {
	var results []*proto.Outcome

	for _, client := range clients {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		outcome, err := client.Result(ctx, &proto.Empty{})
		if err != nil {
			log.Println("Failed to fetch result from a server:", err)
			continue
		}
		results = append(results, outcome)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no results received from any server")
	}

	return results[0], nil
}
