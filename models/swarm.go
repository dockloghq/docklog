package models

type SwarmNode struct {
	ID           string            `json:"id"`
	Hostname     string            `json:"hostname"`
	Role         string            `json:"role"`
	State        string            `json:"state"`
	Availability string            `json:"availability"`
	Manager      bool              `json:"manager"`
	Leader       bool              `json:"leader"`
	EngineVer    string            `json:"engine_version"`
	Addr         string            `json:"addr"`
	CPUs         int64             `json:"cpus"`
	Memory       int64             `json:"memory"`
	Labels       map[string]string `json:"labels,omitempty"`
}

type SwarmService struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Image     string            `json:"image"`
	Mode      string            `json:"mode"`
	Replicas  uint64            `json:"replicas"`
	Running   uint64            `json:"running"`
	Ports     []string          `json:"ports"`
	CreatedAt int64             `json:"created_at"`
	UpdatedAt int64             `json:"updated_at"`
	Labels    map[string]string `json:"labels,omitempty"`
}

type SwarmTask struct {
	ID           string `json:"id"`
	ServiceID    string `json:"service_id"`
	ServiceName  string `json:"service_name"`
	NodeID       string `json:"node_id"`
	NodeHostname string `json:"node_hostname"`
	Slot         int    `json:"slot"`
	Image        string `json:"image"`
	CurrentState string `json:"current_state"`
	DesiredState string `json:"desired_state"`
	Message      string `json:"message,omitempty"`
	Error        string `json:"error,omitempty"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

type SwarmOverview struct {
	Active            bool   `json:"active"`
	LocalNodeState    string `json:"local_node_state"`
	ControlAvailable  bool   `json:"control_available"`
	NodeID            string `json:"node_id"`
	NodesCount        int    `json:"nodes_count"`
	ManagersCount     int    `json:"managers_count"`
	WorkersCount      int    `json:"workers_count"`
	ServicesCount     int    `json:"services_count"`
	TasksCount        int    `json:"tasks_count"`
	RunningTasksCount int    `json:"running_tasks_count"`
}

