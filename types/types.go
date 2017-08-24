package types

import "time"

type PodStatus struct {
	Name   string `json:"name"`
	PodIP  string `json:"pod_ip"`
	NodeIP string `json:"node_ip"`
	Status string `json:"status"`
}

type PodSpec struct {
	Size         int               `json:"size"`
	Version      string            `json:"version,omitempty"`
	NodeSelector map[string]string `json:"node_selector,omitempty"`
}

type Cluster struct {
	Name               string            `json:"name"`
	Pd                 *PodSpec          `json:"pd"`
	Tikv               *PodSpec          `json:"tikv"`
	Tidb               *PodSpec          `json:"tidb"`
	Monitor            *PodSpec          `json:"monitor,omitempty"`
	ServiceType        string            `json:"service_type,omitempty"` // default service type is set at manager startup
	TidbLease          int               `json:"tidb_lease,omitempty"`   // this should be an advanced option
	MonitorReserveDays int               `json:"monitor_reserve_days,omitempty"`
	RootPassword       string            `json:"root_password,omitempty"`
	Labels             map[string]string `json:"labels,omitempty"` // store cluster level meta info

	// response info
	CreatedAt         time.Time   `json:"created_at,omitempty"`
	Initialized       bool        `json:"initialized,omitempty"` // whether initialization password is set
	TidbService       Service     `json:"tidb_service,omitempty"`
	PrometheusService Service     `json:"prometheus_service,omitempty"`
	GrafanaService    Service     `json:"grafana_service,omitempty"`
	PdStatus          []PodStatus `json:"pd_status,omitempty"`
	TidbStatus        []PodStatus `json:"tidb_status,omitempty"`
	TikvStatus        []PodStatus `json:"tikv_status,omitempty"`
}

type Service struct {
	NodeIP       []string `json:"node_ip,omitempty"` // if ServiceType is NodePort or LoadBalancer, NodeIP is all nodes' IP
	NodePort     int      `json:"node_port,omitempty"`
	ClusterIP    string   `json:"cluster_ip,omitempty"`
	ClusterPort  int      `json:"cluster_port,omitempty"`
	ExternalIP   string   `json:"external_ip,omitempty"`   // LoadBalancer IP
	ExternalPort int      `json:"external_port,omitempty"` // LoadBalancer Port
}

type Response struct {
	Action     string  `json:"action"`
	StatusCode int     `json:"status_code"`
	Message    string  `json:"message,omitempty"`
	Payload    Payload `json:"payload,omitempty"`
}

type Payload struct {
	Clusters []*Cluster `json:"clusters,omitempty"`
}
