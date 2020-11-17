package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"net/http"
	"strings"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	natstransport "github.com/go-kit/kit/transport/nats"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

type StringService interface {
	Uppercase(string) (string, error)
	Count(string) int
}

var ErrEmpty = errors.New("empty string")

type stringService struct {
}

func (stringService) Uppercase(s string) (string, error) {
	if s == "" {
		return "", ErrEmpty
	}

	return strings.ToUpper(s), nil
}

func (stringService) Count(s string) int {
	return len(s)
}

type ServiceMiddleware func(StringService) StringService

type uppercaseRequest struct {
	S string `json:"s"`
}

type uppercaseResponse struct {
	V   string `json:"v"`
	Err string `json:"err,omitempty"` // errors don't JSON-marshal, so we use a string
}

type countRequest struct {
	S string `json:"s"`
}

type countResponse struct {
	V int `json:"v"`
}

func makeUppercaseHTTPEndpoint(nc *nats.Conn) endpoint.Endpoint {
	log.WithFields(log.Fields{
		"name": "makeUppercaseHTTPEndpoint",
	}).Info()
	return natstransport.NewPublisher(
		nc,
		"stringsvc.uppercase",
		natstransport.EncodeJSONRequest,
		decodeUppercaseResponse,
	).Endpoint()
}

func decodeUppercaseResponse(_ context.Context, msg *nats.Msg) (interface{}, error) {
	var response uppercaseResponse

	log.WithFields(log.Fields{
		"name": "decodeUppercaseResponse",
	}).Info()
	if err := json.Unmarshal(msg.Data, &response); err != nil {
		return nil, err
	}
	return response, nil
}

func decodeUppercaseHTTPRequest(_ context.Context, req *http.Request) (interface{}, error) {
	log.WithFields(log.Fields{
		"name": "decodeUppercaseHTTPRequest",
	}).Info()
	var request uppercaseRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeUppercaseRequest(_ context.Context, req *nats.Msg) (interface{}, error) {
	log.WithFields(log.Fields{
		"name": "decodeUppercaseRequest",
	}).Info()
	var request uppercaseRequest
	if err := json.Unmarshal(req.Data, &request); err != nil {
		return nil, err
	}
	return request, nil
}

func makeCountHTTPEndpoint(nc *nats.Conn) endpoint.Endpoint {
	log.WithFields(log.Fields{
		"name": "makeCountHTTPEndpoint",
	}).Info()
	return natstransport.NewPublisher(
		nc,
		"stringsvc.count",
		natstransport.EncodeJSONRequest,
		decodeCountResponse,
	).Endpoint()
}

func makeUppercaseEndpoint(svc StringService) endpoint.Endpoint {
	log.WithFields(log.Fields{
		"name": "makeUppercaseEndpoint",
	}).Info()
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(uppercaseRequest)
		v, err := svc.Uppercase(req.S)
		if err != nil {
			return uppercaseResponse{v, err.Error()}, nil
		}
		return uppercaseResponse{V: v, Err: ""}, nil
	}
}

func makeCountEndpoint(svc StringService) endpoint.Endpoint {
	log.WithFields(log.Fields{
		"name": "makeCountEndpoint",
	}).Info()
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(countRequest)
		n := svc.Count(req.S)
		return countResponse{n}, nil
	}
}

func decodeCountResponse(_ context.Context, msg *nats.Msg) (interface{}, error) {
	log.WithFields(log.Fields{
		"name": "decodeCountResponse",
	}).Info()
	var respone countResponse
	if err := json.Unmarshal(msg.Data, &respone); err != nil {
		return nil, err
	}

	return respone, nil
}

func decodeCountHTTPRequest(_ context.Context, req *http.Request) (interface{}, error) {
	log.WithFields(log.Fields{
		"name": "decodeCountHTTPRequest",
	}).Info()
	var request countRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeCountRequest(_ context.Context, req *nats.Msg) (interface{}, error) {
	log.WithFields(log.Fields{
		"name": "decodeCountRequest",
	}).Info()
	var request countRequest
	if err := json.Unmarshal(req.Data, &request); err != nil {
		return nil, err
	}
	return request, nil
}

func main() {
	svc := stringService{}

	natsURL := flag.String("nats-url", nats.DefaultURL, "URL for connecting to NATS")
	listen := flag.String("listen", ":8080", "HTTP Listen Address")
	flag.Parse()

	nc, err := nats.Connect(*natsURL)
	if err != nil {
		log.WithFields(log.Fields{
			"action": "connect",
			"url":    natsURL,
			"err":    err,
		}).Fatal()
	}
	defer nc.Close()

	uppercaseHTTPHandler := httptransport.NewServer(
		makeUppercaseHTTPEndpoint(nc),
		decodeUppercaseHTTPRequest,
		httptransport.EncodeJSONResponse,
	)
	countHTTPHandler := httptransport.NewServer(
		makeCountHTTPEndpoint(nc),
		decodeCountHTTPRequest,
		httptransport.EncodeJSONResponse,
	)

	uppercaseHandler := natstransport.NewSubscriber(
		makeUppercaseEndpoint(svc),
		decodeUppercaseRequest,
		natstransport.EncodeJSONResponse,
	)

	countHandler := natstransport.NewSubscriber(
		makeCountEndpoint(svc),
		decodeCountRequest,
		natstransport.EncodeJSONResponse,
	)

	uSub, err := nc.QueueSubscribe("stringsvc.uppercase", "stringsvc", uppercaseHandler.ServeMsg(nc))
	if err != nil {
		log.WithFields(log.Fields{
			"err":    err,
			"action": "subscribe",
		}).Fatal()
	}
	defer uSub.Unsubscribe()

	cSub, err := nc.QueueSubscribe("stringsvc.count", "stringsvc", countHandler.ServeMsg(nc))
	if err != nil {
		log.WithFields(log.Fields{
			"err":    err,
			"action": "subscribe",
		}).Fatal()
	}
	defer cSub.Unsubscribe()

	http.Handle("/uppercase", uppercaseHTTPHandler)
	http.Handle("/count", countHTTPHandler)
	log.WithFields(log.Fields{
		"event": "Running Server",
		"addr":  *listen,
	}).Info()
	log.Fatal(http.ListenAndServe(*listen, nil))
}
