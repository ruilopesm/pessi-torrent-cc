package main

import (
	"sync"
	"time"
)

type NodeStatistics struct {
	sync.Mutex
	TotalUploaded   uint64
	TotalDownloaded uint64
	nodeMap         map[string][]*DownloadedChunk
}

func NewNodeStatistics() *NodeStatistics {
	return &NodeStatistics{
		nodeMap: make(map[string][]*DownloadedChunk),
	}
}

type DownloadedChunk struct {
	ChunkSize          uint16
	TimestampReceived  time.Time
	TimestampRequested time.Time
}

func (downloadedChunk *DownloadedChunk) getDownloadTime() time.Duration {
	return downloadedChunk.TimestampReceived.Sub(downloadedChunk.TimestampRequested)
}

const RelevantTimeForDownloadSpeedCalculation = 100 * time.Second

func (stats *NodeStatistics) getAverageDownloadSpeed(addr string) float64 {
	stats.Lock()
	defer stats.Unlock()

	downloadedChunks := stats.nodeMap[addr]

	var totalSpeed float64
	var speedCount int

	for _, chunk := range downloadedChunks {
		if time.Now().Sub(chunk.TimestampReceived) < RelevantTimeForDownloadSpeedCalculation {
			downloadTime := chunk.getDownloadTime()
			totalSpeed += float64(chunk.ChunkSize) / downloadTime.Seconds()
			speedCount++
		}
	}

	return totalSpeed / float64(speedCount)
}

func (stats *NodeStatistics) addUploadedBytes(bytes uint64) {
	stats.Lock()
	defer stats.Unlock()
	stats.TotalUploaded += bytes
}

func (stats *NodeStatistics) addDownloadedChunk(addr string, chunkSize uint64, timestampRequested time.Time, timestampReceived time.Time) {
	stats.Lock()
	defer stats.Unlock()

	var val, ok = stats.nodeMap[addr]
	if !ok {
		val = make([]*DownloadedChunk, 0)
	}

	val = append(val, &DownloadedChunk{
		ChunkSize:          uint16(chunkSize),
		TimestampRequested: timestampRequested,
		TimestampReceived:  timestampReceived,
	})

	stats.TotalDownloaded += chunkSize

	stats.nodeMap[addr] = val
}
