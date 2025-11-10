package models

type ContainerStatus string

const (
	StatusCreating ContainerStatus = "creating"
	StatusRunning  ContainerStatus = "running"
	StatusStopped  ContainerStatus = "stopped"
	StatusError    ContainerStatus = "error"
)
