package main

type Profile struct {
	Name            string `json:"name"`
	MaxMemory       string `json:"max_memory"`
	MaxDisk         string `json:"max_disk"`
	MaxNetworkSpeed string `json:"max_network_speed"`
	MaxCPUAllowance string `json:"max_cpu_allowance"`
	CPUPriority     string `json:"cpu_priority"`
	DiskPriority    string `json:"disk_priority"`
}

var (
	BckProfiles     = []byte("profiles")
	DefaultProfiles = []string{"32MB"}
)
