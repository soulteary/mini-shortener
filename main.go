/**
 * Copyright 2022-2025 Su Yang (soulteary)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/soulteary/mini-shortener/internal/version"
)

const (
	DEFAULT_PORT                 = 8901
	RULES_FILE                   = "./rules"
	DEFAULT_RULE                 = `"/ping" => "https://github.com/soulteary/mini-shortener"`
	INFO_FOUND_RULE              = "发现本地配置文件"
	INFO_TRY_CREATE_EXAMPLE_RULE = "尝试创建示例配置文件"
	WARN_RULE_NOT_FOUND          = "没有找到规则文件"
	WARN_RULE_CREATE_FILE        = "尝试创建示例遇到错误"
	WARN_SCAN_RULE_ERR           = "读取规则文件遇到错误"
	WARN_PARSE_RULE_ERR          = "解析规则文件遇到错误"
	ERROR_CAN_NOT_OPEN_RULE      = "读取规则文件遇到错误"
)

type Link struct {
	From string
	To   string
}

var links = make(map[string]string)
var appPort = DEFAULT_PORT

func loadRules(tryToCreateExampleRule bool) (links []Link) {
	if _, err := os.Stat(RULES_FILE); errors.Is(err, os.ErrNotExist) {
		log.Println(WARN_RULE_NOT_FOUND)
		if tryToCreateExampleRule {
			log.Println(INFO_TRY_CREATE_EXAMPLE_RULE)
			err := os.WriteFile(RULES_FILE, []byte(DEFAULT_RULE), 0600)
			if err != nil {
				log.Println(WARN_RULE_CREATE_FILE)
				return links
			}
			return loadRules(false)
		}
		return links
	}
	log.Println(INFO_FOUND_RULE)
	file, err := os.Open(RULES_FILE)
	if err != nil {
		log.Println(ERROR_CAN_NOT_OPEN_RULE)
		log.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		link := parseRules(scanner.Text())
		link.From = strings.TrimSpace(link.From)
		link.To = strings.TrimSpace(link.To)
		if link.From != "" && link.To != "" {
			links = append(links, link)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println(WARN_SCAN_RULE_ERR)
		log.Println(err)
	}
	return links
}

var ruleRegexp = regexp.MustCompile(`"(\/.+)".s*=>.s*"(.+)"`)

func parseRules(input string) (link Link) {
	match := ruleRegexp.FindStringSubmatch(input)
	if len(match) == 3 {
		link.From = match[1]
		link.To = match[2]
		return link
	} else {
		log.Println(WARN_PARSE_RULE_ERR)
		log.Println(input)
	}
	return link
}

var defaults = []byte("Silence is gold")

func route(w http.ResponseWriter, r *http.Request) {
	if redir, ok := links[r.URL.Path]; ok {
		log.Printf("%s => %s\n", r.URL.Path, redir)
		http.Redirect(w, r, redir, http.StatusTemporaryRedirect)
	} else {
		_, err := w.Write(defaults)
		if err != nil {
			log.Printf("程序内部错误 💣")
		}
	}
}

func init() {
	flag.IntVar(&appPort, "port", appPort, "web port")
	flag.Parse()

	portFromEnv := os.Getenv("PORT")
	portEnv, err := strconv.Atoi(portFromEnv)
	if err != nil {
		portEnv = appPort
	}

	userArgs := os.Args[1:]
	if len(userArgs) == 0 {
		if portEnv != DEFAULT_PORT {
			appPort = portEnv
		}
	} else {
		for _, args := range userArgs {
			if !(strings.Contains(args, "--port")) {
				if portEnv != DEFAULT_PORT {
					appPort = portEnv
				}
			}
		}
	}

	for _, link := range loadRules(true) {
		log.Printf("载入规则 %s => %s\n", link.From, link.To)
		links[link.From] = link.To
	}
	log.Println("规则载入完毕 📦")
}

func main() {
	port := strconv.Itoa(appPort)
	http.HandleFunc("/", route)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Println("程序版本：", version.Version)

	srv := &http.Server{
		Addr:              ":" + port,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
	}
	log.Println("服务监听端口：", port)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("程序启动出错: %s\n", err)
		}
	}()

	log.Println("程序已启动完毕 🚀")
	<-ctx.Done()

	stop()
	log.Println("程序正在关闭中，如需立即结束请按 CTRL+C")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("程序强制关闭: ", err)
	}

	log.Println("期待与你的再次相遇 ❤️")
}
