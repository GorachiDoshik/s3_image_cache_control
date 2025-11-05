package app

import (
	"context"
	"fmt"
	"img_cache_control/config"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type App struct {
	Config   *config.Config
	S3Client *s3.Client
}

func NewApp() (*App, error) {

	app := &App{}

	app.Config = config.NewConfig()
	if err := app.Config.LoadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	baseCfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion(app.Config.Storage.Region),
		awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				app.Config.Storage.AccessKey,
				app.Config.Storage.SecretKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load s3 config: %v", err)
	}

	app.S3Client = s3.NewFromConfig(baseCfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.BaseEndpoint = &app.Config.Storage.Endpoint
	})

	return app, nil
}
