package main

import (
	proto "MH_deth/proto"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"math"
	"math/big"
	"math/rand"
	"net"
)

var Logs = flag.Bool("log", false, "determines if you want to see logs")
var p big.Int //Large prime
var q big.Int //large prime such that q % p-1 = 0
var g int64   //rand int with some restrictions
var h big.Int //h = g^a % p
var openedCommitment big.Int
var ourDiceRoll big.Int

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
	if err != nil || err1 != nil {
		log.Printf("failed to unmarshall")
		return &proto.Ack{IsAcknowledged: false}, nil
	}

	if *Logs {
		log.Printf("==== Setup === \np: %v \nq: %v \ng: %v \n", p, q, g)
		fmt.Println("")
	}

	return &proto.Ack{IsAcknowledged: true}, nil
}

//Missleading name, this is where Bob recives Alices Commitment
func (s *server) SendCommitment(ctx context.Context, in *proto.Commitment) (*proto.Ack, error) {
	err := json.Unmarshal(in.H, &h)
	if err != nil {
		log.Printf("failed to unmarshall")
		return &proto.Ack{IsAcknowledged: false}, nil
	}

	if *Logs {
		log.Printf("Received Commitment: \n%v \n", h)
		fmt.Println("")
	}

	die := int64(RollDie())
	ourDiceRoll = *big.NewInt(die)
	return &proto.Ack{IsAcknowledged: true}, nil
}

func (s *server) OpenCommitment(ctx context.Context, in *proto.CommitmentOpener) (*proto.RealRoll, error) {
	if *Logs {
		log.Printf("RECIVED: \n random int: %v \n roll: %v \n", in.RandInt, in.Roll)
		fmt.Println("")
	}

	savedH := h

	//Open commitment by redoing what Alice did
	h.Exp(big.NewInt(g), big.NewInt(in.Roll), nil)
	part1 := big.NewInt(int64(math.Pow(float64(g), float64(in.Roll))))
	part2 := h.Exp(&h, big.NewInt(in.RandInt), nil)
	openedCommitment.Mul(part1, part2)

	log.Printf("The dice roll was: %v \nAlice guessed: %v \n", ourDiceRoll, in.Roll)
	fmt.Println("")
	log.Printf("ALICE HASHED GUESS, REHASHED BY BOB: \n%v \n", openedCommitment)
	fmt.Println("")

	mRoll, _ := ourDiceRoll.MarshalText()

	//Check is Alice open Roll and Random int Matches with what was committed
	fmt.Printf("\nO: %v\nH: %v\nCompareson Equals: %v\n", openedCommitment, savedH, openedCommitment.Cmp(&h))
	if openedCommitment.Cmp(&h) == 0 {
		fmt.Println("Committed H matches with Alice's open message!! YAY!!")
		//Compare what Alice guessed with what was rolled
		if ourDiceRoll.Cmp(big.NewInt(in.Roll)) == 0 {
			return &proto.RealRoll{IsAcknowledged: true, IsGuessed: true, DieRoll: mRoll}, nil
		} else {
			return &proto.RealRoll{IsAcknowledged: true, IsGuessed: false, DieRoll: mRoll}, nil
		}
	} else {
		fmt.Println("Committed H didn't match")
		return &proto.RealRoll{IsAcknowledged: false, IsGuessed: true, DieRoll: mRoll}, nil
	}
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
