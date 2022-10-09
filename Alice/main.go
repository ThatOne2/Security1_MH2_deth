package main

import (
	proto "MH_deth/proto"
	"bufio"
	"context"
	crand "crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"math"
	"math/big"
	"math/rand"
	"os"
	"strings"
)

var Logs = flag.Bool("log", false, "determines if you want to see logs")

var p big.Int            //Large prime such that p = 2 q + 1
var q big.Int            //large prime
var g int64              //rand int with some restrictions
var r int64              //rand int with some restrictions
var h big.Int            //h = g^a % p
var c big.Int            //commitment
var a int64              //what die roll Alice thinks it is
var dieRolledInt big.Int //What was rolled

func MakeParameters() {
	// === find q ===
	qq, _ := crand.Prime(crand.Reader, 1024)
	q = *qq

	// === find p ===
	p.Add(p.Mul(&q, big.NewInt(2)), big.NewInt(1))

	// === find g ===
	g = int64(rand.Intn(100))

	// === find r ===
	r = int64(rand.Intn(100))

	// === find a ===
	a = int64(RollDie())

	// === find h ===
	//Hide what Alice thinks the die roll was
	h.Exp(big.NewInt(g), big.NewInt(a), nil)
	part1 := big.NewInt(int64(math.Pow(float64(g), float64(a))))
	part2 := h.Exp(&h, big.NewInt(r), nil)
	c.Mul(part1, part2)
	//c.Mod(part1.Mul(part1, part2), &p)

	//=== print ===
	if *Logs {
		log.Printf("Setup \n p: %v \nq: %v \ng: %v \na: %v \nh: %v ", p, q, g, a, h)
	}
}

func main() {
	flag.Parse()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("=====")
	fmt.Println("To roll a die write roll and press enter")
	fmt.Println("To end the program press control + c")
	fmt.Println("=====")
	fmt.Println("")

	for scanner.Scan() {
		command := scanner.Text()
		command = strings.TrimSpace(command)
		if command == "roll" {
			MakeCommitment()
		} else {
			fmt.Println("Please write a known command")
		}

	}
}

//Alice and Bob agrees on a group and a base
func SetupConnection() {

	mp, _ := p.MarshalText()
	mq, _ := q.MarshalText()

	req := proto.InitialAgreement{P: mp, Q: mq, G: g}

	connection, err := grpc.Dial(toAddr(8080), grpc.WithInsecure())

	defer connection.Close()
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}

	ctx := context.Background()

	client := proto.NewRollingDieServiceClient(connection)

	response, err := client.SetupAgreements(ctx, &req)
	if err != nil {
		fmt.Println("Connection failed")
		return
	}

	fmt.Printf("Setup is acknowledged: %v\n", response.IsAcknowledged)
	fmt.Println("")
}

func MakeCommitment() {
	MakeParameters()
	SetupConnection()

	mh, _ := c.MarshalText()
	req := proto.Commitment{H: mh}

	//=== print ===
	if *Logs {
		fmt.Println("")
		log.Printf("ALICE BETS ON:\n%v \nHASH H INTO:\n%v \n", a, h)
		fmt.Println("")
	}

	connection, err := grpc.Dial(toAddr(8080), grpc.WithInsecure())

	defer connection.Close()
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}

	ctx := context.Background()

	client := proto.NewRollingDieServiceClient(connection)

	response, err := client.SendCommitment(ctx, &req)
	if err != nil {
		fmt.Println("Connection failed")
		return
	}

	fmt.Printf("Commitment is Acknowledged: %v\n", response.IsAcknowledged)
	fmt.Println("")
	if response.IsAcknowledged {
		OpenCommitment()
	}
}

func OpenCommitment() {
	req := proto.CommitmentOpener{Roll: a, RandInt: r}

	connection, err := grpc.Dial(toAddr(8080), grpc.WithInsecure())

	defer connection.Close()
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}

	ctx := context.Background()

	client := proto.NewRollingDieServiceClient(connection)

	response, err := client.OpenCommitment(ctx, &req)
	if err != nil {
		fmt.Println("Connection failed")
		return
	}

	err = json.Unmarshal(response.DieRoll, &dieRolledInt)
	if err != nil {
		log.Fatalf("Unmarshelling failed")
	}

	fmt.Printf("The Roll: %v\nYou commited to the guess: %v\n", dieRolledInt, a)
	fmt.Println("")
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
