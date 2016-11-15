package main

import (
	"bytes"
	"flag"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/cyberdelia/go-metrics-graphite"
	"github.com/mediocregopher/radix.v2/cluster"
	"github.com/rcrowley/go-metrics"

	"go-jasperlib/jlog"
)

var doneChan chan bool
var keys []string
var redisOpts metrics.Timer
var wg sync.WaitGroup
var numThreads *int
var keySize *int
var expireTime *int
var meterFrequency *int
var redisCluster *cluster.Cluster

func initialize() error {
	doneChan = make(chan bool, *numThreads)
	keys = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n"}
	redisOpts = metrics.NewRegisteredTimer("redisOpts", metrics.DefaultRegistry)

	// initialize graphite metrics recording
	addr, err := net.ResolveTCPAddr("tcp", "qa-scl008-005:2003")
	if err != nil {
		jlog.Warn("Cannot resolve graphite server")
		return err
	}
	go graphite.Graphite(metrics.DefaultRegistry, time.Second*(time.Duration(*meterFrequency)), "rediscluster.shardingtest", addr)

	return nil
}

func stop() {
	for i := 0; i < *numThreads; i++ {
		doneChan <- true
	}
}

func writeRedis() {
	defer wg.Done()

	for {
		// Check if someone told us to stop working
		select {
		case <-doneChan:
			return
		default:
			// continue working
		}
		var buffer bytes.Buffer

		keyDigit := rand.Intn(*keySize)
		for i := 0; i < keyDigit; i++ {
			buffer.WriteString(keys[rand.Intn(len(keys))])
		}

		start := time.Now()
		err := redisCluster.Cmd("incrby", buffer.String(), 1).Err
		if err != nil {
			jlog.Warn(err.Error())
		}
		redisOpts.UpdateSince(start)
		err = redisCluster.Cmd("EXPIRE", buffer.String(), *expireTime).Err
		if err != nil {
			jlog.Warn(err.Error())
		}
	}
}

func wait() {
	// wait for a signal to shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signals
		stop()
	}()

	wg.Wait()
}

func main() {
	numThreads = flag.Int("numThreads", 100, "Set the number of threads")
	keySize = flag.Int("keySize", 3, "Set the key size")
	expireTime = flag.Int("expireTime", 300, "Set the key expire time")
	meterFrequency = flag.Int("meterFrequency", 60, "Set the time interval (sec) of graphite updates")

	flag.Parse()
	err := initialize()
	if err != nil {
		jlog.Warn("initialize failed")
		os.Exit(1)
	}
	redisOpts := cluster.Opts{
		Addr:     "qa-scl007-009:7000",
		PoolSize: *numThreads,
	}
	rc, err := cluster.NewWithOpts(redisOpts)
	if err != nil {
		jlog.Warn(err.Error())
		os.Exit(1)
	}

	redisCluster = rc // avoid variable shadowing
	defer redisCluster.Close()
	wg.Add(*numThreads)
	for i := 0; i < *numThreads; i++ {
		go writeRedis()
	}

	jlog.Info("program started")
	wait()
}
