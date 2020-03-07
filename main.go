package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/m-mizutani/deepalert"
	"github.com/m-mizutani/deepalert/inspector"
	"github.com/m-mizutani/minerva/pkg/api"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type arguments struct {
	Attr      deepalert.Attribute
	SecretArn string // Secret ARN of SecretsManager
}

type secretValues struct {
	MinervaAPIKey   string `json:"minerva_apikey"`
	MinervaEndpoint string `json:"minerva_endpoint"`
	StrixEndpoint   string `json:"strix_endpoint"`
}

const (
	sourceName = "minerva"
)

var (
	logger = logrus.New()
)

func getSecretValues(secretArn string, values interface{}) error {
	// sample: arn:aws:secretsmanager:ap-northeast-1:1234567890:secret:mytest
	arn := strings.Split(secretArn, ":")
	if len(arn) != 7 {
		return errors.New(fmt.Sprintf("Invalid SecretsManager ARN format: %s", secretArn))
	}
	region := arn[3]

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	mgr := secretsmanager.New(ssn)

	result, err := mgr.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretArn),
	})

	if err != nil {
		return errors.Wrapf(err, "Fail to retrieve secret values: %s", secretArn)
	}

	err = json.Unmarshal([]byte(*result.SecretString), values)
	if err != nil {
		return errors.Wrapf(err, "Fail to parse secret values as JSON: %s", secretArn)
	}

	return nil
}

type searchRequest struct {
	Attr *deepalert.Attribute
	secretValues
}

func sendSearchRequest(request searchRequest) (*string, error) {
	url := fmt.Sprintf("%s/api/v1/search", request.MinervaEndpoint)
	now := time.Now().UTC()
	if request.Attr.Timestamp != nil {
		now = *request.Attr.Timestamp
	}

	value := request.Attr.Value
	if request.Attr.Type == deepalert.TypeUserName {
		value = strings.Split(value, "@")[0]
	}

	body := api.ExecSearchRequest{
		Query:         []api.Query{api.Query{Term: value}},
		StartDateTime: now.Add(time.Hour * -2).Format("2006-01-02T15:04:05"),
		EndDateTime:   now.Format("2006-01-02T15:04:05"),
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, errors.Wrapf(err, "Fail to marhsal ExecSearchRequest")
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(raw))
	if err != nil {
		return nil, errors.Wrapf(err, "Fail to create a new request of minerva")
	}

	req.Header.Add("x-api-key", request.MinervaAPIKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "Fail to send request to: %s", url)
	} else if resp == nil {
		return nil, fmt.Errorf("No response data from %s", url)
	}

	logger.WithFields(logrus.Fields{
		"code": resp.StatusCode,
		"url":  url,
	}).Debug("Sent request to minerva")

	var searchResp api.ExecSearchResponse
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "Fail to read data from minerva")
	}
	if err := json.Unmarshal(respBody, &searchResp); err != nil {
		return nil, errors.Wrapf(err, "Fail to unmarshal a result of minerva search API")
	}

	strixURL := fmt.Sprintf("%s/#/search/%s", request.StrixEndpoint, searchResp.SearchID)
	return &strixURL, nil
}

func handler(args arguments) (*deepalert.TaskResult, error) {
	logger.WithField("attr", args.Attr).Debug("Start handler")
	if args.Attr.Type != deepalert.TypeIPAddr &&
		args.Attr.Type != deepalert.TypeUserName &&
		args.Attr.Type != deepalert.TypeDomainName {
		return nil, nil
	}

	var secrets secretValues
	if err := getSecretValues(args.SecretArn, &secrets); err != nil {
		return nil, err
	}

	req := searchRequest{
		Attr:         &args.Attr,
		secretValues: secrets,
	}

	url, err := sendSearchRequest(req)
	if err != nil {
		return nil, err
	}
	if url == nil {
		return nil, fmt.Errorf("No URL from sendSearchRequest")
	}

	now := time.Now().UTC()
	newAttr := deepalert.Attribute{
		Type:      deepalert.TypeURL,
		Value:     *url,
		Key:       fmt.Sprintf("Strix search (%s:%s)", args.Attr.Key, args.Attr.Value),
		Timestamp: &now,
		Context:   deepalert.AttrContexts{deepalert.CtxAdditionalInfo},
	}

	result := deepalert.TaskResult{
		NewAttributes: []deepalert.Attribute{newAttr},
	}
	logger.WithField("result", result).Debug("Exit handler")
	return &result, nil
}

func lambdaHandler(ctx context.Context, attr deepalert.Attribute) (*deepalert.TaskResult, error) {
	args := arguments{
		Attr:      attr,
		SecretArn: os.Getenv("SECRET_ARN"),
	}
	return handler(args)
}

func main() {
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.DebugLevel)

	inspector.Start(inspector.Arguments{
		Handler:         lambdaHandler,
		Author:          sourceName,
		ContentQueueURL: os.Getenv("CONTENT_QUEUE"),
		AttrQueueURL:    os.Getenv("ATTRIBUTE_QUEUE"),
	})
}
