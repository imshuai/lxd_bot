package main

type Node struct {
	Name      string              `json:"name"`
	Address   string              `json:"address"`
	Port      string              `json:"port"`
	LeftQuota int                 `json:"left_quota"`
	Instances map[string]struct{} `json:"instances"`
}
