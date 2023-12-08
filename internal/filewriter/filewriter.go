package filewriter

import (
	"PessiTorrent/internal/logger"
	"PessiTorrent/internal/utils"
	"os"
	"sync"
)

const (
	DownloadsFolder = "downloads"
	Permissions     = 0666
	Flags           = os.O_WRONLY | os.O_CREATE

	WorkerPoolSize = 10
)

type FileWriter struct {
	file        *os.File
	fileName    string
	chunkSize   uint64
	chunksQueue chan Chunk
	stopChannel chan struct{}

	workerWg sync.WaitGroup
}

type Chunk struct {
	index uint16
	data  []uint8
}

func NewFileWriter(fileName string, fileSize uint64) (*FileWriter, error) {
	path := DownloadsFolder + "/" + fileName

	// Create sparse file
	file, err := os.OpenFile(path, Flags, Permissions)
	if err != nil {
		return nil, err
	}
	_, err = file.Seek(int64(fileSize-1), 0)
	if err != nil {
		return nil, err
	}
	_, err = file.Write([]uint8{0})
	if err != nil {
		return nil, err
	}

	return &FileWriter{
		file:        file,
		fileName:    fileName,
		chunkSize:   utils.ChunkSize(fileSize),
		chunksQueue: make(chan Chunk),
		stopChannel: make(chan struct{}),
	}, nil
}

func (fw *FileWriter) EnqueueChunkToWrite(index uint16, data []uint8) {
	fw.chunksQueue <- Chunk{index, data}
}

func (fw *FileWriter) Start() {
	for i := 0; i < WorkerPoolSize; i++ {
		fw.workerWg.Add(1)
		go fw.workerPool()
	}

	<-fw.stopChannel
	fw.workerWg.Wait()
}

func (fw *FileWriter) workerPool() {
	for {
		chunk, ok := <-fw.chunksQueue
		if !ok {
			fw.workerWg.Done()
			return
		}

		fw.writeChunk(chunk)
	}
}

func (fw *FileWriter) writeChunk(chunk Chunk) {
	_, err := fw.file.WriteAt(chunk.data, int64(chunk.index)*int64(fw.chunkSize))
	if err != nil {
		logger.Error("Error writing chunk to file:", err)
	}

	logger.Info("Chunk of index %d written to file %s", chunk.index, fw.fileName)
}

func (fw *FileWriter) Stop() {
	err := fw.file.Close()
	if err != nil {
		logger.Error("Error closing file:", err)
	}

	fw.stopChannel <- struct{}{}
	close(fw.chunksQueue)
	close(fw.stopChannel)
}
