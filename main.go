package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/lucas-clemente/quic-go/http3"
	"github.com/snowmerak/hhh/config"
	"github.com/snowmerak/hhh/loadbalancer"
	"github.com/snowmerak/hhh/ratelimiter"
	"github.com/snowmerak/hhh/system/signal"
)

const VERSION = "0.0.1"

func main() {
	versionFlag := flag.Bool("version", false, "Print version and exit")
	initFlag := flag.String("init", "", "Initialize config file with given name")
	runFlag := flag.String("run", "", "Run the application with the given config file")
	helpFlag := flag.Bool("help", false, "Print help and exit")
	flag.Parse()

	if helpFlag != nil && *helpFlag {
		flag.PrintDefaults()
		os.Exit(0)
	}
	if versionFlag != nil && *versionFlag {
		log.Println(VERSION)
		os.Exit(0)
	}
	if initFlag != nil && *initFlag != "" {
		if err := config.InitAndCreate(*initFlag); err != nil {
			log.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	cf := config.Config{}
	if runFlag != nil && *runFlag != "" {
		if err := config.ReadAndParse(*runFlag, &cf); err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}

	limiter := ratelimiter.New(cf.LimitPerMillisecond, time.Millisecond)
	balancer := loadbalancer.New()
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		counter := 0
		for !limiter.TryTake() && counter < cf.MaxTryCount {
			counter++
			time.Sleep(time.Millisecond)
			runtime.Gosched()
		}
		if counter >= cf.MaxTryCount {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		balancer.Get()
	})

	termSig := signal.NewTerminate()

	for _, l := range cf.Listenings {
		go func(l config.Listening) {
			log.Println("Listening on", l.Address)
			if err := http3.ListenAndServeQUIC(l.Address, l.CertificatePemFile, l.CertificateKeyFile, nil); err != nil {
				log.Println(err)
				os.Exit(1)
			}
		}(l)
	}

	log.Println("Waiting for termination signal")
	<-termSig
}
