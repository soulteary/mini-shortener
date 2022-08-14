package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const rulesFile = "./rules"
const defaultPort = 8901

const (
	WARN_RULE_NOT_FOUND     = "没有找到规则文件"
	WARN_SCAN_RULE_ERR      = "读取规则文件遇到错误"
	WARN_PARSE_RULE_ERR     = "解析规则文件遇到错误"
	ERROR_CAN_NOT_OPEN_RULE = "读取规则文件出错"
)

type Link struct {
	From string
	To   string
}

var links = make(map[string]string)
var appPort = defaultPort

func loadRules() (links []Link) {
	if _, err := os.Stat(rulesFile); errors.Is(err, os.ErrNotExist) {
		fmt.Println(WARN_RULE_NOT_FOUND)
		return links
	}

	file, err := os.Open(rulesFile)
	if err != nil {
		fmt.Println(ERROR_CAN_NOT_OPEN_RULE)
		log.Fatal(err)
	}
	defer file.Close()

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
		fmt.Println(WARN_SCAN_RULE_ERR)
		fmt.Println(err)
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
		fmt.Println(WARN_PARSE_RULE_ERR)
		fmt.Println(input)
	}
	return link
}

var defaults = []byte("Silence is gold")

func route(w http.ResponseWriter, r *http.Request) {
	if redir, ok := links[r.URL.Path]; ok {
		fmt.Printf("[%s] %s => %s\n", time.Now().Format("2006/01/02 15:04:05"), r.URL.Path, redir)
		http.Redirect(w, r, redir, http.StatusTemporaryRedirect)
	} else {
		w.Write(defaults)
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
		fmt.Println(userArgs)
		if portEnv != defaultPort {
			appPort = portEnv
		}
	} else {
		for _, args := range userArgs {
			if !(strings.Contains(args, "--port")) {
				if portEnv != defaultPort {
					appPort = portEnv
				}
			}
		}
	}

	for _, link := range loadRules() {
		fmt.Printf("[Loaded Rule] %s => %s\n", link.From, link.To)
		links[link.From] = link.To
	}
}

func main() {
	port := strconv.Itoa(appPort)
	http.HandleFunc("/", route)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{Addr: ":" + port}
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
