package main

import (
	"bufio"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
)

var (
	errExpired       = errors.New("expired")
	errAboutToExpire = errors.New("expiring in")
	usage            = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	urlFile           = flag.String("urls", "", "path to `file` containing the URLs")
	thresholdDuration = flag.Int("d", 7, "warn of certificate expiration due in `num` days")
)

const maxConcurrentProcs = 4

func main() {
	flag.Parse()
	if *urlFile == "" {
		usage()
		os.Exit(2)
	}

	urls, err := getHosts(*urlFile)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	ch := make(chan struct{}, maxConcurrentProcs)
	for _, url := range urls {
		ch <- struct{}{}
		wg.Add(1)
		go f(url, &wg, ch)
	}
	wg.Wait()
}

func f(host string, wg *sync.WaitGroup, ch chan struct{}) {
	defer wg.Done()
	s, err := checkCertificate(host)
	switch err {
	// expired
	case errExpired:
		fmt.Printf("%v: \033[1;31m%v\033[0m\n", host, err)
	// about to expire
	case errAboutToExpire:
		fmt.Printf("%v: \033[1;33m%v %v\033[0m\n", host, err, s)
	// valid certificate; won't expire soon
	case nil:
		fmt.Printf("%v: \033[1;32mok (will expire in %v)\033[0m\n", host, s)
	// an error occured while checking the certificate
	default:
		fmt.Printf("%v: \033[1;31m%v\033[0m\n", host, err)
	}
	<-ch
}

func checkCertificate(host string) (string, error) {
	dialer := &net.Dialer{Timeout: 15 * time.Second}
	if !strings.Contains(host, ":") {
		host = host + ":443"
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", host, nil)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	if err := conn.Handshake(); err != nil {
		return "", err
	}

	var expiry time.Time
	for _, cert := range conn.ConnectionState().PeerCertificates {
		if cert.IsCA {
			continue
		}
		expiry = cert.NotAfter
		if time.Now().After(expiry) {
			return humanize.Time(expiry), errExpired
		}
		if time.Now().Add(time.Hour * 24 * time.Duration(*thresholdDuration)).After(expiry) {
			return humanize.Time(expiry), errAboutToExpire
		}
	}
	return humanize.Time(expiry), nil
}

// getHosts reads the file containing the URLs and returns
// the hostnames.
//
// It ignores blank lines and lines beginning with `#`.
func getHosts(f string) ([]string, error) {
	file, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hosts := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		txt := scanner.Text()
		txt = strings.TrimSpace(txt)
		if txt == "" || strings.HasPrefix(txt, "#") {
			continue
		}
		hosts = append(hosts, txt)
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}
	return hosts, nil
}
