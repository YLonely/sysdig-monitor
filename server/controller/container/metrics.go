package container

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	metricsNamespace = "sysdig_monitor"
	metricsSubsystem = "container"

	containerIDLabel            = "id"
	syscallTypeLabel            = "syscall_type"
	containerLayerDirLabel      = "layer_dir"
	containerFileLabel          = "file_name"
	containerConnectionSrcIP    = "src_ip"
	containerConnectionDestIP   = "dest_ip"
	containerConnectionSrcPort  = "src_port"
	containerConnectionDestPort = "dest_port"
)

func newGaugeOpts(name, help string) prometheus.GaugeOpts {
	return prometheus.GaugeOpts{Subsystem: metricsSubsystem, Namespace: metricsNamespace, Name: name, Help: help}
}

var (
	// containerIDMetric is a metric tracks the container's name
	containerIDMetric = promauto.NewGaugeVec(
		newGaugeOpts(
			"container_id",
			"The container's id, the value is always 1",
		),
		[]string{containerIDLabel},
	)

	//systemCallCount tracks the total invoke times of different types of system call for every container
	systemCallCount = promauto.NewGaugeVec(
		newGaugeOpts(
			"syscall_total",
			"Total invoke times of different system call",
		),
		[]string{containerIDLabel, syscallTypeLabel},
	)

	//systemCallTotalLatency tracks the total latency of different system call in seconds for every container
	systemCallTotalLatency = promauto.NewGaugeVec(
		newGaugeOpts(
			"system_call_total_latency_seconds",
			"Total latency of the system calls of container",
		),
		[]string{containerIDLabel, syscallTypeLabel},
	)

	// containerLayerDir is metric tracks the layer dir of each container
	containerLayerDir = promauto.NewGaugeVec(
		newGaugeOpts(
			"layer_dir",
			"The dir of the layer of each container, the upper layer's value is 1, the deeper the bigger",
		),
		[]string{containerIDLabel, containerLayerDirLabel},
	)

	// containerLayerFileRead tracks the total bytes a container read from a file which belongs to a layer
	containerLayerFileRead = promauto.NewGaugeVec(
		newGaugeOpts(
			"layer_file_read_bytes",
			"The total bytes a container read from a file which belongs to a layer",
		),
		[]string{containerIDLabel, containerLayerDirLabel, containerFileLabel},
	)

	// containerLayerFileWrite tracks the total bytes a container write to a file which belongs to a layer
	containerLayerFileWrite = promauto.NewGaugeVec(
		newGaugeOpts(
			"layer_file_write_bytes",
			"The total bytes a container write from a file which belongs to a layer",
		),
		[]string{containerIDLabel, containerLayerDirLabel, containerFileLabel},
	)

	// containerActiveConnectionRead tracks the total bytes a container read from a active net connection
	containerActiveConnectionRead = promauto.NewGaugeVec(
		newGaugeOpts(
			"active_connection_read_bytes",
			"The total bytes a container read from a active net connection",
		),
		[]string{containerIDLabel, containerConnectionSrcIP, containerConnectionSrcPort, containerConnectionDestIP, containerConnectionDestPort},
	)

	// containerActiveConnectionWrite tracks the total bytes a container write to a active new connection
	containerActiveConnectionWrite = promauto.NewGaugeVec(
		newGaugeOpts(
			"active_connection_write_bytes",
			"The total bytes a container write to a active net connection",
		),
		[]string{containerIDLabel, containerConnectionSrcIP, containerConnectionSrcPort, containerConnectionDestIP, containerConnectionDestPort},
	)
)
