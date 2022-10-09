package main

import (
	proto "MH_deth/proto"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"math/big"
	"math/rand"
	"net"
)

var Logs = flag.Bool("log", true, "determines if you want to see logs")
var p big.Int //Large prime
var q big.Int //large prime such that q % p-1 = 0
var g int64   //rand int with some restrictions
var h big.Int //h = g^a % p

type server struct {
	proto.UnimplementedRollingDieServiceServer
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", toAddr(8080))

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	proto.RegisterRollingDieServiceServer(s, &server{})

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

//Alice and Bob agrees on a group and a base
func (s *server) SetupAgreements(ctx context.Context, in *proto.InitialAgreement) (*proto.Ack, error) {

	g = in.G
	err := json.Unmarshal(in.P, &p)
	err1 := json.Unmarshal(in.Q, &q)
	err2 := json.Unmarshal(in.H, &h)
	if err != nil || err1 != nil || err2 != nil {
		log.Printf("failed to unmarshall")
		return &proto.Ack{IsAcknowledged: false}, nil
	}

	if *Logs {
		log.Println("================================")
		log.Printf("Setup \n p: %v \n q: %v \n g: %v \n h: %v \n", p, q, g, h)
		log.Println("================================")
	}

	return &proto.Ack{IsAcknowledged: true}, nil
}

//Missleading name, this is where Bob recives Alices Commitment
func (s *server) SendCommitment(ctx context.Context, in *proto.Commitment) (*proto.Ack, error) {
	log.Printf("Received Commitment: %v \n", in.DiceRoll)
	//TODO: finnish
	return &proto.Ack{IsAcknowledged: false}, nil
}

func (s *server) OpenCommitment(ctx context.Context, in *proto.CommitmentOpener) (*proto.Ack, error) {
	log.Printf("Received \n random int: %v \n roll: %v \n", in.RandInt, in.Roll)
	//TODO: finnish
	return &proto.Ack{IsAcknowledged: false}, nil
}

func RollDie() int {
	min := 1
	max := 6

	dieRoll := rand.Intn(max-min) + min

	if *Logs {
		fmt.Println(dieRoll)
	}

	return dieRoll
}

func toAddr(port int32) string {
	return fmt.Sprintf("localhost:%v", port)
}
