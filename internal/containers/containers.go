package containers

import "time"

type Container struct {
	Name      string
	ID        string
	PID       int
	Image     string
	ImageID   string
	StartedAt time.Time
}
