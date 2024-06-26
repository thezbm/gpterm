package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type configType struct {
	url       string
	model     string
	apiKey    string
	httpProxy string
	timeout   int
}

var config configType

func SetUp() {
	viper.SetDefault("profile", "openai")

	viper.SetDefault("openai.url", "https://api.openai.com/v1/chat/completions")
	viper.SetDefault("openai.model", "gpt-3.5-turbo")
	viper.SetDefault("openai.apiKey", "")

	viper.SetDefault("httpProxy", "")
	viper.SetDefault("timeOut", 30)

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	configPath := path.Join(userHomeDir, ".config/gpterm")
	viper.AddConfigPath(configPath)
	viper.SetConfigName("gpterm")
	viper.SetConfigType("toml")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := os.MkdirAll(configPath, 0755); err != nil {
				log.Fatal(err)
			}
			if err := viper.SafeWriteConfig(); err != nil {
				panic(fmt.Errorf("fatal error config file: %w", err))
			}
			fmt.Printf("No config file found.\nA default config file is created at [%s].\nPlease set the apiKey (and optionally the httpProxy) before chatting with bot.\n", configPath)
			os.Exit(0)
		} else {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	profile := viper.GetString("profile")
	config = configType{
		url:       viper.GetString(profile + ".url"),
		model:     viper.GetString(profile + ".model"),
		apiKey:    viper.GetString(profile + ".apiKey"),
		httpProxy: viper.GetString("httpProxy"),
		timeout:   viper.GetInt("timeout"),
	}
	if config.apiKey == "" {
		fmt.Println("Please set the apiKey before chatting with bot.")
		os.Exit(0)
	}
}

const (
	apiReqBody = `{
        "model": "%s",
        "messages": %s
    }`
)

type messageType struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

var (
	messages []messageType
)

func Ask(input string) string {
	messages = append(messages, messageType{"user", input})
	messagesBytes, err := json.Marshal(messages)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", config.url, bytes.NewBuffer([]byte(fmt.Sprintf(apiReqBody, config.model, messagesBytes))))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.apiKey))

	proxyUrl, err := url.Parse(config.httpProxy)
	if err != nil {
		log.Fatal(err)
	}
	res, err := (&http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
		Timeout: time.Duration(config.timeout * int(time.Second)),
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

	output := apiRes.Choices[0].Message.Content
	messages = append(messages, messageType{"assistant", output})

	return output
}

func GetModel() string { return config.model }
