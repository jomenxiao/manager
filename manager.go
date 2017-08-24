package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/GregoryIan/manager/types"
)

const (
	create = "create"
	query  = "query"
	delete = "delete"
)

var defaultCount = 3

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
	flag.IntVar(&tidbCount, "tidb-count", defaultCount, "tidb pod count")
	flag.IntVar(&tikvCount, "tikv-count", defaultCount, "tikv pod count")
	flag.IntVar(&pdCount, "pd-count", defaultCount, "pd pod count")
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
		fmt.Printf("create cluster %s at %s\n", cluster, url)
		xpost(url, cluster)
	case query:
		if name != "" {
			url = fmt.Sprintf("%s/%s", url, name)
		}
		xget(url)
	case delete:
		checkDeleteCluster()
		url = fmt.Sprintf("%s/%s", url, name)
		fmt.Printf("delete cluster %s at %s\n", name, url)
		xdelete(url)
	default:
		fatalf("unsupport cmd %s", cmd)
	}
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
}

func checkDeleteCluster() {
	if name == "" {
		fatal("lack of cluster name")
	}
}

func createClusterRequest() *types.Cluster {
	cluster := &types.Cluster{
		Name: name,
	}

	cluster.Pd = &types.PodSpec{
		Version: pdVersion,
		Size:    pdCount,
	}

	cluster.Tikv = &types.PodSpec{
		Version: tikvVersion,
		Size:    tikvCount,
	}

	cluster.Tidb = &types.PodSpec{
		Version: tidbVersion,
		Size:    tidbCount,
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

func xget(url string) {
	res, err := http.Get(url)
	if err != nil {
		fatal(err.Error())
	}
	content, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		fatal(err.Error())
	}
	fmt.Printf("%s\n", content)
}

func xpost(url string, body []byte) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		fatalf("create cluster request error %v", err)
	}
	request(req)
}

func xdelete(url string) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fatalf("delete cluster request error %v", err)
	}
	request(req)
}

func request(req *http.Request) {
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		fatalf("issue request error %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("success")
	}

	bodyByte, _ := ioutil.ReadAll(resp.Body)
	fatalf("fail to request %s", bodyByte)
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
