package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"math"
	"time"
)

type CloudWatchService struct {
	Client cloudwatchiface.CloudWatchAPI
}

type metricMapping struct {
	Name      string
	Statistic string
}

func NewCloudWatchService(sess *session.Session) CloudWatchService {
	return CloudWatchService{
		Client: cloudwatch.New(sess),
	}
}

func (cw *CloudWatchService) GetCDNMetricsForDistribution(distributionId string) ([]*cloudwatch.GetMetricStatisticsOutput, error) {
	metricNames := []metricMapping{
		{Name: "Requests", Statistic: "Sum"},
		{Name: "BytesDownloaded", Statistic: "Sum"},
		{Name: "BytesUploaded", Statistic: "Sum"},
		{Name: "4xxErrorRate", Statistic: "Average"},
		{Name: "5xxErrorRate", Statistic: "Average"},
		{Name: "TotalErrorRate", Statistic: "Average"},
	}

	var outputs []*cloudwatch.GetMetricStatisticsOutput
	endTime := time.Now()
	startTime := time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC)
	period := calculatePeriod(startTime, endTime)

	for _, mapping := range metricNames {

		unit := aws.String("None")
		if mapping.Statistic == "Average" {
			unit = aws.String("Percent")
		}

		input := cloudwatch.GetMetricStatisticsInput{
			Dimensions: []*cloudwatch.Dimension{
				{
					Name:  aws.String("DistributionId"),
					Value: aws.String(distributionId),
				},
				{
					Name:  aws.String("Region"),
					Value: aws.String("Global"),
				},
			},
			EndTime:    aws.Time(endTime),
			MetricName: aws.String(mapping.Name),
			Namespace:  aws.String("AWS/CloudFront"),
			Period:     aws.Int64(period),
			StartTime:  aws.Time(startTime),
			Statistics: []*string{aws.String(mapping.Statistic)},
			Unit:       unit,
		}

		output, err := cw.Client.GetMetricStatistics(&input)

		if err != nil {
			return nil, err
		}

		outputs = append(outputs, output)
	}

	return outputs, nil
}

// calculates the number hours between the two times, in seconds
func calculatePeriod(startTime time.Time, endTime time.Time) int64 {
	numSeconds := endTime.Sub(startTime).Seconds()
	return int64(
		math.Round(
			numSeconds + (60 - math.Mod(numSeconds, 60)),
		),
	)
}
