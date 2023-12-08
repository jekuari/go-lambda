package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type ColorEvent struct {
	Color struct {
		Hex string `json:"hex"`
	} `json:"color"`
}

func HandleRequest(ctx context.Context, event ColorEvent) (string, error) {
	// Parse hex color to RGB
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
	err = uploadToS3(buf)
	if err != nil {
		return "", err
	}

	return "Image successfully generated and uploaded to S3", nil
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

func uploadToS3(buf []byte) error {
	credentials := credentials.NewEnvCredentials()
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials,
		Region:      aws.String("us-east-1"),
	})
	if err != nil {
		return fmt.Errorf("failed to create AWS session: %v", err)
	}

	imgRead := bytes.NewReader(buf)

	// Create an S3 client
	s3Client := s3.New(sess)

	// Upload the image to S3
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String("colors"),
		Key:         aws.String("generated_image.png"),
		Body:        aws.ReadSeekCloser(imgRead),
		ACL:         aws.String("public-read"), // Set appropriate ACL
		ContentType: aws.String("image/png"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload image to S3: %v", err)
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
