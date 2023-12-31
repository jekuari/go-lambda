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
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type APIGWResponse struct {
	IsBase64Encoded bool              `json:"isBase64Encoded"`
	StatusCode      int               `json:"statusCode"`
	Headers         map[string]string `json:"headers"`
	Body            string            `json:"body"`
}

var empty APIGWResponse

func HandleRequest(ctx context.Context, event map[string]interface{}) (APIGWResponse, error) {
	// Parse hex color to RGB

	fmt.Println(event["body"], ctx)
	fmt.Printf("%v", reflect.TypeOf(event["body"]))

	body := event["body"].(string)
	decoder := json.NewDecoder(strings.NewReader(body))

	decoded := make(map[string]string)
	err := decoder.Decode(&decoded)

	if err != nil {
		panic(err)
	}

	rgbColor, err := parseHexColor(decoded["color"])
	if err != nil {
		return empty, err
	}

	// Create a 32x32 image with the specified color
	img := CreateImage(rgbColor)

	// Encode the image to PNG
	var buf bytes.Buffer
	err = png.Encode(&buf, img)

	if err != nil {
		return empty, err
	}

	// Upload the image to S3
	fileName, err := uploadToS3(&ctx, buf.Bytes())
	if err != nil {
		return empty, err
	}

	// convert res to json

	res := fmt.Sprintf("https://color.s3.amazonaws.com/%s", fileName)

	finalRes := APIGWResponse{
		IsBase64Encoded: false,
		StatusCode:      200,
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
		Body: res,
	}

	return finalRes, nil
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
