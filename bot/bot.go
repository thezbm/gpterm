package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	apiURL     = "https://api.openai.com/v1/chat/completions"
	apiReqBody = `{
        "model": "%s",
        "messages": [
            %s
        ]
    }`
)

type messageStruct struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func SetUp() {
	viper.SetDefault("apiKey", "")
	viper.SetDefault("model", "gpt-3.5-turbo")
	viper.SetDefault("httpProxy", "")
	viper.SetDefault("timeOut", 30)
	viper.SetConfigName("gpterm")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := viper.SafeWriteConfig(); err != nil {
				panic(fmt.Errorf("fatal error config file: %w", err))
			}
			fmt.Println("No config file found. A default config file is created. Please set the apiKey and the httpProxy.")
		} else {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}
}

func Ask(input string) string {
	userMessage, err := json.Marshal(messageStruct{"user", input})
	if err != nil {
		log.Fatal(err)
	}
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer([]byte(fmt.Sprintf(apiReqBody, viper.GetString("model"), userMessage))))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", viper.GetString("apiKey")))

	proxyUrl, err := url.Parse(viper.GetString("httpProxy"))
	if err != nil {
		log.Fatal(err)
	}
	res, err := (&http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
		Timeout: time.Duration(viper.GetInt("timeOut") * int(time.Second)),
	}).Do(req)
	if err != nil {
		log.Fatal(err)
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, resBody)
	}

	var apiRes struct {
		Choices [1]struct {
			Message struct {
				Content string
			}
		}
	}
	dec := json.NewDecoder(strings.NewReader(string(resBody)))
	err = dec.Decode(&apiRes)
	if err != nil {
		log.Fatal(err)
	}

	return apiRes.Choices[0].Message.Content
}

func GetModel() string { return viper.GetString("model") }
