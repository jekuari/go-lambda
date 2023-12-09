package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Event struct {
	Body struct {
		Color string `json:"color"`
	} `json:"body"`
}

func HandleRequest(ctx context.Context, event map[string]interface{}) (string, error) {
	// Parse hex color to RGB
	println("event", event)
	fmt.Println(event, ctx)
	rgbColor, err := parseHexColor(event["Body"].(map[string]interface{})["color"].(string))
	if err != nil {
		return "", err
	}

	// Create a 32x32 image with the specified color
	img := CreateImage(rgbColor)

	// Encode the image to PNG
	var buf bytes.Buffer
	err = png.Encode(&buf, img)

	if err != nil {
		return "", err
	}

	// Upload the image to S3
	fileName, err := uploadToS3(&ctx, buf.Bytes())
	if err != nil {
		return "", err
	}

	// convert res to json

	res := fmt.Sprintf("https://color.s3.amazonaws.com/%s", fileName)

	return res, nil
}

func parseHexColor(hex string) (color.RGBA, error) {
	// Parse hex color to RGB
	var rgbColor color.RGBA
	_, err := fmt.Sscanf(hex, "#%02x%02x%02x", &rgbColor.R, &rgbColor.G, &rgbColor.B)
	if err != nil {
		return color.RGBA{}, fmt.Errorf("failed to parse color: %v", err)
	}
	rgbColor.A = 255 // Alpha channel

	return rgbColor, nil
}

func CreateImage(rgbColor color.RGBA) image.Image {
	// Create a 32x32 image with the specified color
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	draw.Draw(img, img.Bounds(), &image.Uniform{rgbColor}, image.Point{}, draw.Src)

	return img
}

func uploadToS3(ctx *context.Context, buf []byte) (string, error) {
	cfg, err := config.LoadDefaultConfig(*ctx)

	if err != nil {
		return "", fmt.Errorf("failed to load AWS SDK config: %v", err)
	}

	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.Region = "us-east-1"
	})

	bucketName := "color"
	fileName := fmt.Sprintf("%s.png", time.Now().Format("2006-01-02-15-04-05-000000000"))
	imageReader := bytes.NewReader(buf)

	params := &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &fileName,
		Body:   imageReader,
	}

	_, err = client.PutObject(*ctx, params)

	// Upload the image to S3

	if err != nil {
		return "", fmt.Errorf("failed to upload image to S3: %v", err)
	}

	return fileName, nil
}

func main() {
	lambda.Start(HandleRequest)
}
