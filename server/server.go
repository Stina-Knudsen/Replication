package main

import (
	"time"

	proto "Replication/grpc"
)

const auctionDuration = 100 * time.Second

func main() {

}

type AuctionServer struct {
	proto.UnimplementedAuctionServerServer
}
