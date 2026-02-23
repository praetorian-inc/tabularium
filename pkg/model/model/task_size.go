package model

type TaskSize struct {
	CPU    int `json:"cpu"`    // vCPU units: 256, 512, 1024, 2048, 4096, 8192, 16384
	Memory int `json:"memory"` // MB: 512, 1024, 2048, 4096, 8192, 16384, ...
}

var DefaultTaskSize = TaskSize{CPU: 1024, Memory: 2048}
