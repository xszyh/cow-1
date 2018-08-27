package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/bigtan/cow/proxy"

	"github.com/elazarl/goproxy"
	"github.com/shadowsocks/go-shadowsocks2/core"
)

var settings struct {
	Server     string `json:"server_addr"`
	Local      string `json:"local_addr"`
	ServerPort int    `json:"server_port"`
	HTTPPort   int    `json:"http_port"`
	SocksPort  int    `json:"socks_port"`
	Password   string `json:"password"`
	Method     string `json:"method"`
	Verbose    bool   `json:"verbose"`
}

func main() {

	var flags struct {
		Config  string
		Blocked string
	}

	var tlsClientSkipVerify = &tls.Config{InsecureSkipVerify: true}

	flag.StringVar(&flags.Config, "c", "config.json", "file name of config file")
	flag.StringVar(&flags.Blocked, "b", "blocked.txt", "file name of blocked file")
	flag.Parse()

	// init logger
	logger := log.New(os.Stderr, "cow: ", log.LstdFlags|log.Lshortfile)

	// open config file
	configFile, err := os.Open(flags.Config)
	defer configFile.Close()

	if err != nil {
		logger.Fatal(err)
	}

	// parse config file
	jsonParser := json.NewDecoder(configFile)

	if err = jsonParser.Decode(&settings); err != nil {
		logger.Fatal("Parsing config file failed", err.Error())
	}

	// setup shadowsocks client
	ciph, err := core.PickCipher(settings.Method, nil, settings.Password)

	if err != nil {
		logger.Fatal("No such cipher", err.Error())
	}

	// setup local route
	server := fmt.Sprintf("%s:%d", settings.Server, settings.ServerPort)
	localSOCKS := fmt.Sprintf("%s:%d", settings.Local, settings.SocksPort)
	localHTTP := fmt.Sprintf("%s:%d", settings.Local, settings.HTTPPort)

	go socksLocal(localSOCKS, server, ciph.StreamConn)

	// setup http proxy
	prxy := goproxy.NewProxyHttpServer()

	prxy.Logger = logger
	prxy.Verbose = settings.Verbose
	prxy.ConnectDial = proxy.Direct.Dial

	// proxy bypass to localSOCKS
	bypass, _ := proxy.SOCKS5("tcp", localSOCKS, nil, proxy.Direct)

	// setup per_host
	perHost := proxy.NewPerHost(proxy.Direct, bypass)

	prxy.Tr = &http.Transport{Dial: perHost.Dial, Proxy: nil, TLSClientConfig: tlsClientSkipVerify}

	file, err := os.Open(flags.Blocked)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		rule := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(rule, "#") && len(rule) > 0 {
			perHost.AddFromString(rule)
		}
	}

	logger.Fatal(http.ListenAndServe(localHTTP, prxy))
}
