package grpc

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	porterGRPC "github.com/MChorfa/porter/pkg/grpc/proto/porter/v1"
)

type server struct {
	porterGRPC.UnimplementedPorterServiceServer
}

func NewServer() *server {
	return &server{}
}

func (s *server) Install(ctx context.Context, in *porterGRPC.IntallRequest) (*porterGRPC.IntallResponse, error) {
	return &porterGRPC.IntallResponse{Message: in.body}, nil
}
func main() {
	// Create a listener on TCP port
	lis, err := net.Listen("tcp", ":7777")
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Create a gRPC server object
	s := grpc.NewServer()
	// Attach the Greeter service to the server
	porterGRPC.RegisterPorterServiceServer(s, &server{})
	// Serve gRPC server
	log.Println("Serving gRPC on 0.0.0.0:7777")
	go func() {
		log.Fatalln(s.Serve(lis))
	}()

	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	conn, err := grpc.DialContext(
		context.Background(),
		"0.0.0.0:7777",
		grpc.WithBlock(),
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()
	// Register Greeter
	err = porterGRPC.RegisterPorterServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	gwServer := &http.Server{
		Addr:    ":7779",
		Handler: gwmux,
	}

	log.Println("Serving gRPC-Gateway on http://0.0.0.0:7779")
	log.Fatalln(gwServer.ListenAndServe())
}
