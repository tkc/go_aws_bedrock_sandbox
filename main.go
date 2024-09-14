package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
)

// Define a struct to hold the AWS credentials from the YAML file
type AWSCredentials struct {
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	ModelID         string `yaml:"model_id"`
}

// Function to load AWS credentials from a YAML file
func loadAWSCredentials(filename string) (*AWSCredentials, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var creds AWSCredentials
	err = yaml.Unmarshal(data, &creds)
	if err != nil {
		return nil, err
	}
	return &creds, nil
}

func main() {
	// Load AWS credentials from YAML file
	conf, err := loadAWSCredentials("conf.yaml")
	if err != nil {
		fmt.Println("Failed to load AWS credentials:", err)
		return
	}

	// Configuration
	region := "ap-northeast-1"

	// Setting the endpoint and path
	endpoint := fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com", region)
	path := fmt.Sprintf("/model/%s/invoke", conf.ModelID)
	url := endpoint + path

	// Preparing the payload (specified format)
	payload := map[string]interface{}{
		"anthropic_version": "bedrock-2023-05-31",
		"max_tokens":        1000,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "Hello, Claude",
					},
				},
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Failed to convert payload to JSON:", err)
		return
	}

	// Creating the request body
	body := bytes.NewReader(payloadBytes)

	// Creating the HTTP request
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		fmt.Println("Failed to Request:", err)
		return
	}

	// Setting the headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Amz-Target", "BedrockRuntime.InvokeModel")

	// Creating the AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
		Credentials: credentials.NewStaticCredentials(
			conf.AccessKeyID,
			conf.SecretAccessKey,
			"",
		),
	})
	if err != nil {
		fmt.Println("Failed to create AWS session:", err)
		return
	}

	// Resetting the body reader to reuse for signing
	var bodyReader io.ReadSeeker = bytes.NewReader(payloadBytes)

	// Signing the request
	signer := v4.NewSigner(sess.Config.Credentials)
	_, err = signer.Sign(req, bodyReader, "bedrock", region, time.Now())
	if err != nil {
		fmt.Println("Failed to sign request:", err)
		return
	}

	// Sending the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to send request:", err)
		return
	}
	defer resp.Body.Close()

	// Reading the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error occurred:", err)
		return
	}

	// Processing the response
	if resp.StatusCode == 200 {
		fmt.Println("", string(respBody))
	} else {
		fmt.Printf("error: %d\n", resp.StatusCode)
		fmt.Println(string(respBody))
	}
}
