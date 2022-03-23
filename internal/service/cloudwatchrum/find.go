package cloudwatchrum

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchrum"
)

func FindAppMonitorByName(ctx context.Context, conn *cloudwatchrum.CloudWatchRUM, name string) (*cloudwatchrum.AppMonitor, error) {
	input := cloudwatchrum.GetAppMonitorInput{
		Name: aws.String(name),
	}

	output, err := conn.GetAppMonitorWithContext(ctx, &input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.AppMonitor, nil
}
