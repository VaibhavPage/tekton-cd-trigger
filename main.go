package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"github.com/VaibhavPage/tekton-cd-trigger/proto"
	"github.com/ghodss/yaml"
	cdpipeline "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1alpha1"
	"google.golang.org/grpc"
	"k8s.io/client-go/rest"
)

type TektonCDTrigger struct {
	pipelineClient v1alpha1.TektonV1alpha1Interface
}

type TriggerResource struct {
	URL string `json:"url"`
}

// FetchResource fetches the resource to be triggered.
func (t *TektonCDTrigger) FetchResource(ctx context.Context, in *proto.FetchResourceRequest) (*proto.FetchResourceResponse, error) {
	var resource *TriggerResource
	if err := yaml.Unmarshal(in.Resource, &resource); err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Get(resource.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &proto.FetchResourceResponse{
		Resource: content,
	}, nil
}

// Execute executes the requested trigger resource.
func (t *TektonCDTrigger) Execute(ctx context.Context, in *proto.ExecuteRequest) (*proto.ExecuteResponse, error) {
	var p *cdpipeline.PipelineRun
	if err := yaml.Unmarshal(in.Resource, &p); err != nil {
		return nil, err
	}

	response, err := t.pipelineClient.PipelineRuns(p.Namespace).Create(p)
	if err != nil {
		return nil, err
	}

	fmt.Printf("pipeline run %s successfully executed\n", response.Name)
	return &proto.ExecuteResponse{
		Response: []byte("success"),
	}, nil
}

// ApplyPolicy applies policies on the trigger execution result.
func (t *TektonCDTrigger) ApplyPolicy(ctx context.Context, in *proto.ApplyPolicyRequest) (*proto.ApplyPolicyResponse, error) {
	return &proto.ApplyPolicyResponse{
		Success: true,
		Message: "success",
	}, nil
}

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "9000"
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		panic(err)
	}
	srv := grpc.NewServer()

	trigger := &TektonCDTrigger{
		pipelineClient: v1alpha1.NewForConfigOrDie(config),
	}

	proto.RegisterTriggerServer(srv, trigger)

	fmt.Println("starting trigger server")

	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
