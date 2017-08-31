package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	create = "create"
	query  = "query"
	delete = "delete"
)

var defaultPDCount = 1
var defaultTiDBCount = 1
var defaultTiKVCount = 5
var maxWaitCount = 60

var (
	cloudManagerAddr string
	cmd              string
	tidbVersion      string
	tidbCount        int
	tikvVersion      string
	tikvCount        int
	pdVersion        string
	pdCount          int
	name             string
)

func init() {
	flag.StringVar(&cloudManagerAddr, "cloud-manager-addr", "", "the addr of cloud manager")
	flag.StringVar(&cmd, "cmd", query, "the command against to cloud manager")
	flag.StringVar(&name, "name", "", "tidb image version")
	flag.StringVar(&tidbVersion, "tidb-version", "", "tidb image version")
	flag.StringVar(&tikvVersion, "tikv-version", "", "tikv image version")
	flag.StringVar(&pdVersion, "pd-version", "", "pd image version")
	flag.IntVar(&tidbCount, "tidb-count", defaultTiDBCount, "tidb pod count")
	flag.IntVar(&tikvCount, "tikv-count", defaultTiKVCount, "tikv pod count")
	flag.IntVar(&pdCount, "pd-count", defaultPDCount, "pd pod count")
}

func main() {
	flag.Parse()
	if cloudManagerAddr == "" {
		fatal("lack of cloud-manager-addr")
	}

	url := cloudManagerAddr + "/pingcap.com/api/v1/clusters"
	switch cmd {
	case create:
		checkCreateClusterParameter()
		cluster := createCluster()
		xpost(url, cluster)
		url = fmt.Sprintf("%s/%s", url, name)
		getClusterAccessInfo(url)
	case query:
		if name != "" {
			url = fmt.Sprintf("%s/%s", url, name)
		}
		resp := xget(url)
		for _, cluster := range resp.Payload.Clusters {
			fmt.Printf("%+v\n", cluster)
		}
	case delete:
		checkDeleteCluster()
		url = fmt.Sprintf("%s/%s", url, name)
		fmt.Printf("delete cluster %s at %s\n", name, url)
		xdelete(url)
	default:
		fatalf("unsupport cmd %s", cmd)
	}
}

func getClusterAccessInfo(url string) {
	var clusters []*Cluster
	var cluster *Cluster
	var index int
	for ; index < maxWaitCount; index++ {
		response := xget(url)
		clusters = response.Payload.Clusters
		if len(clusters) == 0 {
			fatalf("don't find cluster")
		}

		cluster = clusters[0]

		if !checkPodStatus(cluster.PdStatus, cluster.Pd.Size) {
			time.Sleep(10 * time.Second)
			continue
		}

		if !checkPodStatus(cluster.TikvStatus, cluster.Tikv.Size) {
			time.Sleep(10 * time.Second)
			continue
		}

		if !checkPodStatus(cluster.TidbStatus, cluster.Tidb.Size) {
			time.Sleep(10 * time.Second)
			continue
		}

		if len(cluster.TidbService.NodeIP) > 0 {
			break
		}

		time.Sleep(10 * time.Second)
		continue
	}

	// wait bootstarp
	time.Sleep(time.Minute)
	if index >= maxWaitCount {
		//xdelete(url)
		fatalf("can't wait cluster %s", url)
	}
	host := waitTiDBOK(cluster, url)
	fmt.Println("host:", host)
	fmt.Println("port:", cluster.TidbService.NodePort)
}

func checkPodStatus(status []PodStatus, size int) bool {
	running := 0
	for _, s := range status {
		if s.Status == "Running" {
			running++
		}
	}

	return running >= size
}

func waitTiDBOK(cluster *Cluster, url string) string {
	length := len(cluster.TidbService.NodeIP)

	var (
		index = 0
		host  = "127.0.0.1"
		port  = cluster.TidbService.NodePort
	)
	var err error
	for ; index < maxWaitCount; index++ {
		selectedNode := rand.Int() % length
		host = cluster.TidbService.NodeIP[selectedNode]

		err = connectTiDB(host, port)
		if err != nil {
			fmt.Printf("connection tidb %s:%d error %v, continue\n", host, port, err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	if index >= maxWaitCount {
		// xdelete(url)
		fatalf("can't wait cluster %s, error %v", url, err)
	}

	return host
}

func connectTiDB(host string, port int) error {
	dsn := fmt.Sprintf("root@tcp(%s:%d)/mysql?charset=utf8&timeout=3s", host, port)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	rs, err := db.Query("SELECT count(*) FROM mysql.tidb")
	if err != nil {
		return err
	}
	defer rs.Close()
	var tidbCount int64
	for rs.Next() {
		err := rs.Scan(&tidbCount)
		if err != nil {
			return err
		}
		break
	}

	return nil
}

func checkCreateClusterParameter() {
	if name == "" {
		fatal("lack of cluster name")
	}

	if tidbVersion == "" {
		fatal("lack of tidb version")
	}

	if tikvVersion == "" {
		fatal("lack of tikv version")
	}

	if pdVersion == "" {
		fatal("lack of pd version")
	}

	if tidbCount < 3 {
		tidbCount = 3
	}
}

func checkDeleteCluster() {
	if name == "" {
		fatal("lack of cluster name")
	}
}

func createClusterRequest() *Cluster {
	cluster := &Cluster{
		Name:               name,
		TidbLease:          5,
		MonitorReserveDays: 14,
	}

	cluster.Pd = &PodSpec{
		Version: pdVersion,
		Size:    pdCount,
	}

	cluster.Tikv = &PodSpec{
		Version: tikvVersion,
		Size:    tikvCount,
	}

	cluster.Tidb = &PodSpec{
		Version: tidbVersion,
		Size:    tidbCount,
	}

	cluster.Monitor = &PodSpec{
		Version: "4.2.0,v1.5.2,v0.3.1",
		Size:    1,
	}
	return cluster
}

func createCluster() []byte {
	cluster := createClusterRequest()
	body, err := json.Marshal(cluster)
	if err != nil {
		fatalf("create cluster error %v", err)
	}
	return body
}

func xget(url string) *Response {
	res, err := http.Get(url)
	if err != nil {
		fatal(err.Error())
	}
	content, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		fatal(err.Error())
	}

	response := &Response{}
	err = json.Unmarshal(content, response)
	if err != nil {
		fatalf("unmarshal error %v", err)
	}

	if response.StatusCode != 200 {
		fatalf("fail to request %v", response)
	}

	return response
}

func xpost(url string, body []byte) *Response {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		fatalf("create cluster request error %v", err)
	}
	return request(req)
}

func xdelete(url string) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fatalf("delete cluster request error %v", err)
	}
	request(req)
}

func request(req *http.Request) *Response {
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		fatalf("issue request error %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fatalf("fail to request: status code %d", resp.StatusCode)
	}

	bodyByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fatalf("fail to read body %v", err)
	}

	response := &Response{}
	err = json.Unmarshal(bodyByte, response)
	if err != nil {
		fatalf("unmarshal error %v", err)
	}

	if response.StatusCode != 200 {
		fatalf("fail to request %v", response)
	}

	return response
}

func fatal(message string) {
	fmt.Println(message)
	os.Exit(-1)
}

func fatalf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	fmt.Println("")
	os.Exit(-1)
}
