package main

import (
	"net/http"
	"log"
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"encoding/json"
	"io/ioutil"
	"time"
)

var addr = kingpin.Flag("address", "Address to serve at").Default(":80").String()

func errorHandler(w http.ResponseWriter, message string, code int) {
	// Marshal the response
	responseJson, _ := json.Marshal(&errorResponse{
		Success: false,
		Message: message,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(responseJson)
}

// getS3Client returns a new S3 client
func getS3Client(creds *credentials.Credentials, region string, endpoint string) (*s3.S3, error) {
	s3Config := &aws.Config{
		Credentials: creds,
		Region:      aws.String(region),
		Endpoint:    aws.String(endpoint),
	}

	sess, err := session.NewSession(s3Config)
	if err != nil {
		return nil, err
	}

	return s3.New(sess), nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorHandler(w, "unknown method", 400)
		return
	}

	// Unmarshal the payload
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorHandler(w, err.Error(), 500)
		return
	}
	var req request
	err = json.Unmarshal(body, &req)
	if err != nil {
		errorHandler(w, err.Error(), 400)
		return
	}

	// Get the s3 client
	s3Client, err := getS3Client(credentials.NewStaticCredentials(req.ID, req.Secret, ""), req.Region, req.Endpoint)
	if err != nil {
		errorHandler(w, err.Error(), 500)
		return
	}

	// Create the request
	putRequest, _ := s3Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket:   aws.String(req.Bucket),
		ACL:      aws.String(req.ACL),
		Key:      aws.String(req.Key),
		Metadata: req.Metadata,
	})

	// Parse the expiration
	expiry, err := time.ParseDuration(req.Expiry)
	if err != nil {
		errorHandler(w, err.Error(), 400)
		return
	}

	// Presign the request
	url, headers, err := putRequest.PresignRequest(expiry)
	if err != nil {
		errorHandler(w, err.Error(), 500)
		return
	}

	// Marshal the response
	responseJson, err := json.Marshal(&response{
		Success: true,
		Uri:     url,
		Method:  putRequest.HTTPRequest.Method,
		Headers: headers,
	})
	if err != nil {
		errorHandler(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJson)
}

func main() {
	kingpin.Parse()
	log.Printf("Serving at %v", *addr)
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Panic(err)
	}
}

type request struct {
	ID       string             `json:"id"`
	Secret   string             `json:"secret"`
	Region   string             `json:"region"`
	Endpoint string             `json:"endpoint"`
	Bucket   string             `json:"bucket"`
	Key      string             `json:"key"`
	ACL      string             `json:"acl"`
	Expiry   string             `json:"expiry"`
	Metadata map[string]*string `json:"metadata"`
}

type errorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type response struct {
	Success bool                `json:"success"`
	Uri     string              `json:"uri"`
	Method  string              `json:"method"`
	Headers map[string][]string `json:"headers"`
}
