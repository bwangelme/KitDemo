package addtransport

import (
	"context"
	"fmt"
	"kitdemo/addsvc/pb"
	"kitdemo/addsvc/pkg/addservice"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	"google.golang.org/grpc"

	"kitdemo/addsvc/pkg/addendpoint"

	grpctransport "github.com/go-kit/kit/transport/grpc"
)

type grpcServer struct {
	sum    grpctransport.Handler
	concat grpctransport.Handler
}

func NewGRPCServer(endpoints addendpoint.Set, logger log.Logger) pb.AddServer {
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}
	return &grpcServer{
		sum: grpctransport.NewServer(
			endpoints.SumEndpoint,
			decodeGRPCSumRequest,
			encodeGRPCSumResponse,
			options...,
		),
		concat: grpctransport.NewServer(
			endpoints.ConcatEndpoint,
			decodeGRPCConcatRequest,
			encodeGRPCConcatResponse,
			options...,
		),
	}
}

func NewGRPCClient(conn *grpc.ClientConn, logger log.Logger) addservice.Service {
	var options []grpctransport.ClientOption

	var sumEndpoint endpoint.Endpoint
	{
		sumEndpoint = grpctransport.NewClient(
			conn,
			"pb.Add",
			"Sum",
			encodeGRPCSumRequest,
			decodeGRPCSumResponse,
			pb.SumReply{},
			options...,
		).Endpoint()
	}

	var concatEndpoint endpoint.Endpoint
	{
		concatEndpoint = grpctransport.NewClient(
			conn,
			"pb.Add",
			"Concat",
			encodeGRPCConcatRequest,
			decodeGRPCConcatResponse,
			pb.ConcatReply{},
			options...,
		).Endpoint()
	}

	return addendpoint.Set{
		SumEndpoint:    sumEndpoint,
		ConcatEndpoint: concatEndpoint,
	}
}

func (s *grpcServer) Sum(ctx context.Context, req *pb.SumRequest) (*pb.SumReply, error) {
	_, rep, err := s.sum.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	// 这里是将 rep 转换成了指针类型
	return rep.(*pb.SumReply), nil
}
func (s *grpcServer) Concat(ctx context.Context, req *pb.ConcatRequest) (*pb.ConcatReply, error) {
	_, rep, err := s.concat.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	// 这里是将 rep 转换成了指针类型
	return rep.(*pb.ConcatReply), nil
}

func decodeGRPCConcatRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.ConcatRequest)
	return addendpoint.ConcatRequest{A: req.A, B: req.B}, nil
}

func decodeGRPCSumRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.SumRequest)
	return addendpoint.SumRequest{A: int(req.A), B: int(req.B)}, nil
}

func encodeGRPCSumRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(addendpoint.SumRequest)
	return &pb.SumRequest{A: int64(req.A), B: int64(req.B)}, nil
}

func decodeGRPCSumResponse(_ context.Context, grpcResp interface{}) (interface{}, error) {
	resp := grpcResp.(*pb.SumReply)
	return addendpoint.SumResponse{V: int(resp.V), Err: str2err(resp.Err)}, nil
}

func encodeGRPCSumResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(addendpoint.SumResponse)
	return &pb.SumReply{V: int64(resp.V), Err: err2str(resp.Err)}, nil
}

func encodeGRPCConcatResponse(_ context.Context, grpcResp interface{}) (interface{}, error) {
	resp := grpcResp.(addendpoint.ConcatResponse)
	return &pb.ConcatReply{V: resp.V, Err: err2str(resp.Err)}, nil
}

func encodeGRPCConcatRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(addendpoint.ConcatRequest)
	return &pb.ConcatRequest{A: req.A, B: req.B}, nil
}

func decodeGRPCConcatResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*pb.ConcatReply)
	return addendpoint.ConcatResponse{V: resp.V, Err: str2err(resp.Err)}, nil
}

func err2str(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func str2err(str string) error {
	if str == "" {
		return nil
	}
	return fmt.Errorf("%s", str)
}
