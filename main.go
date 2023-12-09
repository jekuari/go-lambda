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

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
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
	cfg.Region = "us-east-1"
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS SDK config: %v", err)
	}
	imgRead := bytes.NewReader(buf)

	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	_, err = s3Client.CreateSession(*ctx, &s3.CreateSessionInput{
		Bucket: aws.String("colors"),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	// Upload the image to S3
	res, err := s3Client.PutObject(*ctx, &s3.PutObjectInput{
		Bucket:      aws.String("colors"),
		Key:         aws.String("generated_image.png"),
		Body:        aws.ReadSeekCloser(imgRead),
		ContentType: aws.String("image/png"),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to upload image to S3: %v", err)
	}

	return res, nil
}

func main() {
	lambda.Start(HandleRequest)
}
