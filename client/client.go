package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	proto "Replication/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var bidder string

func main() {
	file, e := os.OpenFile("../auction_log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if e != nil {
		log.Fatalf("failed to open log file: %v", e)
	}
	defer file.Close()

	log.SetOutput(file)
	// --------------------------------------
	// List of server addresses (replicas)
	servers := []string{":50051", ":50052", ":50053"}

	// Choose a server to connect to
	var conn *grpc.ClientConn
	var err error
	var client proto.AuctionServerClient

	log.Println("Starting client...")
	for _, server := range servers {
		log.Printf("Client attempting to connect to server: %v", server)
		conn, err = grpc.Dial(server, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			client = proto.NewAuctionServerClient(conn)
			log.Printf("Client connected to server: %v", server)
			break
		} else {
			log.Printf("Client failed to connect to server %v: %v", server, err)
		}

	}
	if err != nil {
		log.Fatalf("Client failed to connect to any server: %v", err)
	}
	defer conn.Close()

	// --------------------------------------

	fmt.Print("Enter your username: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	bidder = scanner.Text()

	for {
		fmt.Println("Type 'bid' to bid in the auction, or type 'result' to view the result :))")
		scanner.Scan()
		command := scanner.Text()
		command = strings.TrimSpace(command)

		if command == ("bid") {
			fmt.Println("Type the amount you want to bid")
			scanner.Scan()
			amount, err := strconv.ParseInt(scanner.Text(), 10, 32)
			if err != nil {
				log.Print("failed to parse amount: %d", amount)
			}
			Bid(context.Background(), client, bidder, int32(amount))
		} else if command == "result" {
			Result(context.Background(), client)
		} else {
			fmt.Println("Unidentified command")
		}
	}
}

func Bid(ctx context.Context, client proto.AuctionServerClient, bidder string, amount int32) {
	log.Printf("Sending bid request: User %s, Amount %d", bidder, amount)

	req := &proto.Amount{
		Amount: amount,
		Bidder: bidder,
	}

	// Send the request to the server
	resp, err := client.Bid(ctx, req)
	if err != nil {
		log.Printf("Error while bidding: %v", err)
		return
	}

	// Handle the response
	log.Printf("Response from server: %s", resp.Ack)
	if resp.Ack == "success" {
		fmt.Println("Your bid was accepted!")
	} else {
		fmt.Println("Your bid was rejected:", resp.Ack)
	}
}

func Result(ctx context.Context, client proto.AuctionServerClient) {
	log.Println("Fetching auction result...")

	req := &proto.Empty{}

	// Send the request to the server
	resp, err := client.Result(ctx, req)
	if err != nil {
		log.Printf("Error while fetching result: %v", err)
		fmt.Println("Failed to fetch the auction result. Please try again later.")
		return
	}

	// Handle the response
	log.Printf("Auction result received: %s, Highest Bid: %d", resp.Result, resp.HighestBid)
	fmt.Printf("Auction Status: %s\n", resp.Result)
	fmt.Printf("Current Highest Bid: %d\n", resp.HighestBid)
}
