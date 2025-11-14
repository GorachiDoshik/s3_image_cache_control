package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type CacheControlUpdater struct {
	App *App
}

const (
	BATCH_SIZE      = 50
	CACHE_CONTROL   = "max-age=2592000"
	PREFIX_DEV_DIR  = "cache_control_test"
	PREFIX_PROD_DIR = "iblock"
)

func (cu *CacheControlUpdater) Run() error {

	ctx := context.Background()
	client := cu.App.S3Client
	bucket := cu.App.Config.Storage.Bucket

	failedLog, err := os.OpenFile("logs/fail_updated_img.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open failed log file: %v", err)
	}
	defer failedLog.Close()

	appLog, err := os.OpenFile("logs/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open app log file: %v", err)
	}
	defer appLog.Close()

	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(PREFIX_PROD_DIR),
	})

	pageNum := 0
	totalProcessed := 0
	batchProcessed := 0

	for paginator.HasMorePages() {

		pageNum++
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("error fetching S3 page: %w", err)
		}

		log.Printf("---- Processing S3 Page #%d ----", pageNum)

		for _, obj := range page.Contents {

			key := *obj.Key
			if !strings.HasSuffix(key, ".webp") &&
				!strings.HasSuffix(key, ".jpg") &&
				!strings.HasSuffix(key, ".jpeg") &&
				!strings.HasSuffix(key, ".png") {
				continue
			}

			head, err := client.HeadObject(ctx, &s3.HeadObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(key),
			})
			if err != nil {
				log.Printf("failed to read metadata for %s: %v", key, err)
				fmt.Fprintln(failedLog, key)
				continue
			}

			log.Printf("[%d] %s", totalProcessed+1, key)
			// log.Printf("  Content-Type: %s", aws.ToString(head.ContentType))
			// log.Printf("  Cache-Control: %s", aws.ToString(head.CacheControl))
			// log.Printf("  Last-Modified: %v", head.LastModified)
			// log.Printf("-----------------------------------------")

			_, err = client.CopyObject(ctx, &s3.CopyObjectInput{
				Bucket:            aws.String(bucket),
				CopySource:        aws.String(bucket + "/" + key),
				Key:               aws.String(key),
				CacheControl:      aws.String(CACHE_CONTROL),
				ContentType:       head.ContentType,
				MetadataDirective: "REPLACE",
			})
			if err != nil {
				log.Printf("Failed to update Cache-Control for %s: %v", key, err)
				fmt.Fprintln(failedLog, key)
				continue
			}

			batchProcessed++
			totalProcessed++

			if batchProcessed >= BATCH_SIZE {
				progressMsg := fmt.Sprintf("Processed batch: %d images, total processed: %d\n", batchProcessed, totalProcessed)
				fmt.Fprint(appLog, progressMsg)

				log.Print("🌈 " + progressMsg)
				// log.Printf("🌈 batch of %d completed, moving to next batch...", BATCH_SIZE)
				batchProcessed = 0
			}
		}
	}

	fmt.Fprintf(appLog, "finished processing. Total updated: %d\n", totalProcessed)

	log.Printf("done proccess! Total updated: %d objects.", totalProcessed)
	return nil
}
