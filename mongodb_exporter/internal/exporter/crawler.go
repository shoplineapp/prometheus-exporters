package exporter

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	dac "github.com/xinsnake/go-http-digest-auth-client"

	"github.com/shoplineapp/go-app/plugins/env"
	"github.com/shoplineapp/go-app/plugins/logger"
	"github.com/sirupsen/logrus"
)

type Crawler struct {
	env    *env.Env
	logger *logger.Logger
	events *Events
}

const (
	MONGO_ATLAS_API_ROOT = "cloud.mongodb.com/api/atlas/v1.0"
	MONGO_ATLAS_LOG_API  = "groups/%s/clusters/%s/logs/mongodb.gz"
)

func (c Crawler) ReadGzippedData(data []byte) string {
	reader := bytes.NewReader(data)
	gzreader, e1 := gzip.NewReader(reader)
	if e1 != nil {
		fmt.Println(e1)
	}

	output, e2 := ioutil.ReadAll(gzreader)
	if e2 != nil {
		fmt.Println(e2)
	}

	return string(output)
}

func (c *Crawler) DownloadLogs(cluster string, server string) {
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			c.logger.WithFields(logrus.Fields{"cluster": cluster, "server": server, "err": err}).Error("Failed to get logs from Mongo Atlas")
			c.events.Publish(EVENT_LOGS_DOWNLOADED, cluster, server, "")
		}
	}()

	api := fmt.Sprintf(MONGO_ATLAS_LOG_API, c.env.GetEnv("MONGO_ATLAS_GROUP_ID"), server)

	sinceStr := c.env.GetEnv("CRAWLER_SINCE_TIME")
	since, _ := time.ParseDuration(sinceStr)

	start := time.Now().Add(-1 * since)
	c.logger.WithFields(logrus.Fields{"cluster": cluster, "server": server, "since": start.Unix()}).Info("Getting logs from Mongo Atlas")

	url := fmt.Sprintf("https://%s/%s?startDate=%d", MONGO_ATLAS_API_ROOT, api, start.Unix())

	// Request with digest
	t := dac.NewTransport(c.env.GetEnv("MONGO_ATLAS_PUBLIC_KEY"), c.env.GetEnv("MONGO_ATLAS_PRIVATE_KEY"))
	req, rErr := http.NewRequest("GET", url, nil)
	if rErr != nil {
		c.logger.WithFields(logrus.Fields{"error": rErr}).Error("Failed to construct request to Mongo Atlas")
	}
	req.Header.Set("Accept", "application/gzip")
	resp, tErr := t.RoundTrip(req)
	if tErr != nil {
		c.logger.WithFields(logrus.Fields{"error": tErr}).Error("Failed to request from Mongo Atlas")
	}

	defer resp.Body.Close()
	data, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		c.logger.WithFields(logrus.Fields{"error": readErr}).Error("Unable to read response from Mongo Atlas")
	} else {
		logs := c.ReadGzippedData(data)
		c.logger.WithFields(logrus.Fields{"length": len(logs)}).Info("Mongo logs updated")
		// ioutil.WriteFile("output.txt", []byte(logs), 0644)

		c.events.Publish(EVENT_LOGS_DOWNLOADED, cluster, server, logs)
	}
}

func (c *Crawler) Run() {
	var wg sync.WaitGroup
	cluster := c.env.GetEnv("MONGO_ATLAS_CLUSTER_NAME")
	servers := strings.Split(c.env.GetEnv("MONGO_ATLAS_CLUSTER_IDS"), ",")
	for _, server := range servers {
		go func(server string) {
			c.logger.WithFields(logrus.Fields{"cluster": cluster, "server": server}).Info("Start downloading logs from Mongo Atlas")
			c.DownloadLogs(cluster, server)
		}(server)
	}
	for i := 0; i < len(servers); i++ {
		server := servers[i]
		wg.Add(1)
		go func(cluster string, srv string) {
			c.events.SubscribeOnce(fmt.Sprintf(EVENT_LOGS_ENTRIES_STORED_BY_SERVER, cluster, srv), func() {
				wg.Done()
			})
		}(cluster, server)
	}
	wg.Wait()

	c.logger.WithFields(logrus.Fields{"cluster": cluster}).Info("Finished downloading logs from Mongo Atlas")
	c.events.Publish(EVENT_LOGS_SERVERS_RECEIVED)
}

func (c Crawler) Listen() {
	intervalStr := c.env.GetEnv("CRAWLER_INTERVAL_TIME")
	interval, _ := time.ParseDuration(intervalStr)

	ticker := time.NewTicker(interval)
	quit := make(chan struct{})
	go func() {
		c.Run()
		for {
			select {
			case <-ticker.C:
				c.Run()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	c.logger.Info("Crawler listening for logs")
}

func NewCrawler(env *env.Env, logger *logger.Logger, events *Events) *Crawler {
	c := &Crawler{
		env:    env,
		events: events,
		logger: logger,
	}
	return c
}
