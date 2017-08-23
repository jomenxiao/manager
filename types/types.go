package types

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
	Name string   `json:"name"`
	Pd   *PodSpec `json:"pd"`
	Tikv *PodSpec `json:"tikv"`
	Tidb *PodSpec `json:"tidb"`
}
