package main

import (
	proto "MH_deth/proto"
	"bufio"
	"context"
	crand "crypto/rand"
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

var Logs = flag.Bool("log", true, "determines if you want to see logs")

var p big.Int //Large prime such that p = 2 q + 1
var q big.Int //large prime
var g int64   //rand int with some restrictions
var h big.Int //h = g^a % p
var a int64   //what die roll Alice thinks it is

func MakeParameters() {
	// === find q ===
	qq, _ := crand.Prime(crand.Reader, 1024)
	q = *qq

	// === find p ===
	p.Add(p.Mul(&q, big.NewInt(2)), big.NewInt(1))

	// === find g ===
	g = int64(rand.Intn(100))

	// === find a ===
	a = int64(RollDie())

	// === find h ===
	h.Mod(big.NewInt(int64(math.Pow(float64(g), float64(a)))), &p)

	//=== print ===
	if *Logs {
		log.Printf("Setup \n p: %v \n q: %v \n g: %v \n h: %v \n a: %v \n ", p, q, g, h, a)
	}
}

func main() {
	flag.Parse()
	MakeParameters()

	SetupConnection()
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("To roll a die write roll and press enter")
	fmt.Println("To end the program press control + c")

	for scanner.Scan() {
		UserName := scanner.Text()
		UserName = strings.TrimSpace(UserName)
		if len(UserName) < 1 {

		} else {
			fmt.Println("Please write a known command")
		}

	}
}

//Alice and Bob agrees on a group and a base
func SetupConnection() {

	mp, _ := p.MarshalText()
	mq, _ := q.MarshalText()
	mh, _ := h.MarshalText()

	req := proto.InitialAgreement{P: mp, Q: mq, G: g, H: mh}

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
}

func MakeCommitment() {
	//min := 1
	//max := 9999

	//r := rand.Intn(max-min) + min

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
