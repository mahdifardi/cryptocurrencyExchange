package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	ServerPort string `json:"ServerPort"`

	EthServerAddress string `json:"EthServerAddress"`

	BtccHostAddress string `json:"BtcHostAddress"`
	BtcPort         string `json:"BtcPort"`
	BtcWallet       string `json:"BtcWallet"`
	BtcEndpoint     string `json:"BtcEndpoint"`

	BtcUser    string `json:"BtccUser"`
	BtcPass    string `json:"BtccPass"`
	BtccParams string `json:"BtcParams"`

	BtcUser1Address string `json:"BtcUser1Address"`
	BtcUser2Address string `json:"BtcUser2Address"`
	BtcUser3Address string `json:"BtcUser3Address"`

	EthUser1Address string `json:"EthUser1Address"`
	EthUser2Address string `json:"EthUser2Address"`
	EthUser3Address string `json:"EthUser3Address"`

	User1ID int64 `json:"User1ID"`
	User2ID int64 `json:"User2ID"`
	User3ID int64 `json:"User3ID"`
}

func LoadConfig(path string) (Config, error) {
	var config Config

	file, err := os.Open(path)
	if err != nil {
		return config, err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	return config, err
}
