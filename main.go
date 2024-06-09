package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/viper"
)

const (
	apiURL     = "https://api.openai.com/v1/chat/completions"
	apiReqBody = `{
        "model": "%s",
        "messages": [
            {
                "role": "system",
                "content": "Make sure all your responses are in Markdown format."
            },
            {
                "role": "user",
                "content": "%s"
            }
        ]
    }`
)

func main() {
	viper.SetDefault("apiKey", "")
	viper.SetDefault("httpProxy", "")
	viper.SetDefault("timeOut", 30)
	viper.SetDefault("model", "gpt-4-turbo")
	viper.SetConfigName("gpterm")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println(viper.SafeWriteConfig())
			fmt.Println("No config file found. A default config file is created. Please set the apiKey and the httpProxy.")
		} else {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	apiKey := viper.GetString("apiKey")
	httpProxy := viper.GetString("httpProxy")

	input, _, err := bufio.NewReader(os.Stdin).ReadLine()
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer([]byte(fmt.Sprintf(apiReqBody, viper.GetString("model"), input))))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	proxyUrl, err := url.Parse(httpProxy)
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

	renderedOutput, err := glamour.Render(apiRes.Choices[0].Message.Content, "dark")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", renderedOutput)
}
