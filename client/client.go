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

var userID string

func main() {

	// --------------------------------------
	// List of server addresses (replicas)
	servers := []string{":50051", ":50052", ":50053"}

	// Choose a server to connect to
	var conn *grpc.ClientConn
	var err error
	var client proto.AuctionServerClient

	for _, server := range servers {
		conn, err = grpc.Dial(server, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			client = proto.NewAuctionServerClient(conn)
			log.Printf("Connected to server: %v", server)
			break
		}
		log.Printf("Failed to connect to server %v: %v", server, err)
	}
	if err != nil {
		log.Fatalf("Failed to connect to any server: %v", err)
	}
	defer conn.Close()

	// --------------------------------------

	fmt.Print("Enter your username: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	userID = scanner.Text()

	for {
		fmt.Println("Type 'bid' to bid in the auction, or type 'result' to view the result :))")
		command := scanner.Text()
		command = strings.TrimSpace(command)

		if command == ("bid") {
			fmt.Println("Type the amount you want to bid")
			amount, err := strconv.ParseInt(scanner.Text(), 10, 32)
			if err != nil {
				log.Print("failed to parse amount: %d", amount)
			}

			Bid(context.Background(), client, userID, int32(amount))
			break
		} else if command == "result" {
			Result(context.Background(), client)
			break
		} else {
			fmt.Println("Unidentified command")
		}
	}
}

func Bid(ctx context.Context, client proto.AuctionServerClient, userID string, amount int32) {

}

func Result(ctx context.Context, client proto.AuctionServerClient) {

}
