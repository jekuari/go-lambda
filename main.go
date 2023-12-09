package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type ColorEvent struct {
	Color struct {
		Hex string `json:"hex"`
	} `json:"color"`
}

func HandleRequest(ctx context.Context, event ColorEvent) (string, error) {
	// Parse hex color to RGB
	fmt.Println(ctx, event)
	rgbColor, err := parseHexColor(event.Color.Hex)
	if err != nil {
		return "", err
	}

	// Create a 32x32 image with the specified color
	img := createImage(rgbColor)

	// Encode the image to PNG
	var buf []byte
	err = encodePNG(&buf, img)
	if err != nil {
		return "", err
	}

	fmt.Println("v2, ", buf)

	// Upload the image to S3
	res, err := uploadToS3(&ctx, buf)
	if err != nil {
		return "", err
	}

	// convert res to json
	encodedRes, err := json.Marshal(res)
	if err != nil {
		return "", err
	}

	return fmt.Sprint(encodedRes), nil
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

func createImage(rgbColor color.RGBA) image.Image {
	// Create a 32x32 image with the specified color
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	draw.Draw(img, img.Bounds(), &image.Uniform{rgbColor}, image.Point{}, draw.Src)

	return img
}

func encodePNG(buf *[]byte, img image.Image) error {
	// Encode the image to PNG
	err := png.Encode(os.Stdout, img)
	if err != nil {
		return fmt.Errorf("failed to encode image to PNG: %v", err)
	}

	return nil
}

func uploadToS3(ctx *context.Context, buf []byte) (*s3.PutObjectOutput, error) {
	cfg, err := config.LoadDefaultConfig(*ctx)
	fmt.Println("v1")

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS SDK config: %v", err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
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

	res, err := client.PutObject(*ctx, params)

	// Upload the image to S3

	if err != nil {
		return nil, fmt.Errorf("failed to upload image to S3: %v", err)
	}

	return res, nil
}

func main() {
	lambda.Start(HandleRequest)
}
