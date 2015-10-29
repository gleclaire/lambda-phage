// +build -ignore

package main

import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/aws/aws-sdk-go/service/lambda"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/aws/aws-sdk-go/service/s3"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws/awserr"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/tj/go-debug"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/spf13/cobra"

var eventSources []EventSource

func init() {

	esCmd := &cobra.Command{
		Use:   "event-sources",
		Short: "sets up event sources as set in your config file",
		RunE:  setupEventSources,
	}

	cmds = append(cmds, esCmd)
}

type EventSource interface {
	Setup() error
}

// sets up them event sources
func setupEventSources(c *cobra.Command, _ []string) error {
	for _, es := range cfg.EventSources {
		var drv EventSourceSetup
		switch es.Type {

		}
	}
}
