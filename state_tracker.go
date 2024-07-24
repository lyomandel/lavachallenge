package main


import (
	"io/ioutil"
	"flag"
	"fmt"
	"context"
	"time"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc"
	"log"
	"encoding/json"
	pb "cosmossdk.io/api/cosmos/base/tendermint/v1beta1"
)

const (
	MaxCallRecvMsgSize  = 1024 * 1024 * 32 // setting receive size to 32mb instead of 4mb default
)

var (
	serverAddr = flag.String("addr", "localhost:50051", "The server address in the format of host:port")
)

type BlockData struct {
	Height	int64 	`json:"height"`
	Hash	[]byte 	`json:"hash"`
}

type TestResult struct {
	Result []BlockData `json:"test_result"`
}

func getLatestBlockData(client pb.ServiceClient) BlockData {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	latest_block, err := client.GetLatestBlock(ctx, &pb.GetLatestBlockRequest{})
	if err != nil {
		log.Fatalf("fail to GetLatestBlock: %v", err)
	}
	block_data := BlockData{
		Height: latest_block.Block.Header.Height,
		Hash: 	latest_block.BlockId.Hash,
	}
	return block_data
}

func readAmountOfLatestBlockData(amount_of_blocks int, client pb.ServiceClient) []BlockData {
	block_data_array := make([]BlockData, amount_of_blocks)
	var last_block_height int64
	last_block_height = 0
	for i := range amount_of_blocks {
		fmt.Printf("Reading block number: %d out of %d. \n", i+1, amount_of_blocks)
		block_data_array[i] = getLatestBlockData(client)
		for last_block_height == block_data_array[i].Height {
			block_data_array[i] = getLatestBlockData(client)
		}
		last_block_height = block_data_array[i].Height
		
		block_data_json, _ := json.Marshal(block_data_array[i])
		fmt.Println(string(block_data_json))
	}
	return block_data_array
} 

func writeBlockDataArrayToFile(block_data_array []BlockData) {
	test_result := TestResult{
		Result: block_data_array,
	}
	block_data_json, _ := json.Marshal(test_result)
	err := ioutil.WriteFile("output.json", block_data_json, 0644)
	if err != nil {
		log.Fatalf("fail to write file: %v", err)
	}
	fmt.Printf("%+v\n", string(block_data_json))
}


func getLavaProxyClient() pb.ServiceClient {
	opts := []grpc.DialOption{
        grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), 
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize)),
	}
	conn, err := grpc.NewClient(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	return pb.NewServiceClient(conn)
}

func main(){
	client := getLavaProxyClient()
	block_data_array := readAmountOfLatestBlockData(5,client)
	writeBlockDataArrayToFile(block_data_array)
}