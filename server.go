
package main
import (
	"fmt"
	"flag"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/credentials"
	"crypto/tls"
	"log"
	"net"
	pb "cosmossdk.io/api/cosmos/base/tendermint/v1beta1"
)

const (
	MaxCallRecvMsgSize  = 1024 * 1024 * 32 // setting receive size to 32mb instead of 4mb default
)

var (
	port       = flag.Int("port", 50051, "The server port")
	lavaServerAddr = flag.String("addr", "lav1.grpc.lava.build:443", "The server address in the format of host:port")
)

type tenderMintServiceServer struct {
	pb.UnimplementedServiceServer
	proxyClient pb.ServiceClient
}

func (s *tenderMintServiceServer) GetNodeInfo(ctx context.Context, request *pb.GetNodeInfoRequest) (*pb.GetNodeInfoResponse, error) {
	return s.proxyClient.GetNodeInfo(ctx, request)
}
func (s *tenderMintServiceServer) GetSyncing(ctx context.Context, request *pb.GetSyncingRequest) (*pb.GetSyncingResponse, error) {
	return s.proxyClient.GetSyncing(ctx, request)
}
func (s *tenderMintServiceServer) GetLatestBlock(ctx context.Context, request *pb.GetLatestBlockRequest) (*pb.GetLatestBlockResponse, error) {
	return s.proxyClient.GetLatestBlock(ctx, request)
}
func (s *tenderMintServiceServer) GetBlockByHeight(ctx context.Context, request *pb.GetBlockByHeightRequest) (*pb.GetBlockByHeightResponse, error) {
	return s.proxyClient.GetBlockByHeight(ctx, request)
}
func (s *tenderMintServiceServer) GetLatestValidatorSet(ctx context.Context, request *pb.GetLatestValidatorSetRequest) (*pb.GetLatestValidatorSetResponse, error) {
	return s.proxyClient.GetLatestValidatorSet(ctx, request)
}
func (s *tenderMintServiceServer) GetValidatorSetByHeight(ctx context.Context,request *pb.GetValidatorSetByHeightRequest) (*pb.GetValidatorSetByHeightResponse, error) {
	return s.proxyClient.GetValidatorSetByHeight(ctx, request)
}
func (s *tenderMintServiceServer) ABCIQuery(ctx context.Context, request *pb.ABCIQueryRequest) (*pb.ABCIQueryResponse, error) {
	return s.proxyClient.ABCIQuery(ctx, request)
}
func (s *tenderMintServiceServer) mustEmbedUnimplementedServiceServer() {}

func getLavaGRPCClient(ctx context.Context) pb.ServiceClient {
	var opts []grpc.DialOption
	// opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	var tlsConf tls.Config
	tlsConf.InsecureSkipVerify = true // Allows self-signed certificates
	credentials := credentials.NewTLS(&tlsConf)
	opts = append(opts, grpc.WithTransportCredentials(credentials))
	opts = append(opts, grpc.WithBlock(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize)))
	opts = append(opts, grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
		return net.Dial("tcp", *lavaServerAddr)
	}))
	opts = append(opts, grpc.WithDefaultCallOptions(
		grpc.UseCompressor(gzip.Name), // Use gzip compression for provider consumer communication
	))
	conn, err := grpc.DialContext(ctx, *lavaServerAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial rpc client: %v", err)
	}
	return pb.NewServiceClient(conn)
}

func newServer(client pb.ServiceClient) *tenderMintServiceServer {
	s := &tenderMintServiceServer{}
	s.proxyClient = client
	return s
}

func setupTendermintProxy(ctx context.Context) *grpc.Server{
	lavaClient := getLavaGRPCClient(ctx)
	fmt.Println("Test connection by reading latest block.")
	lavaClient.GetLatestBlock(ctx, &pb.GetLatestBlockRequest{})
	fmt.Println("Tendermint connection established successfully.")
	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterServiceServer(grpcServer, newServer(lavaClient))
    reflection.Register(grpcServer)
	return grpcServer
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	ctx := context.Background()
	grpcServer := setupTendermintProxy(ctx)
	fmt.Println("Starting to serve!")
	grpcServer.Serve(lis)
}
