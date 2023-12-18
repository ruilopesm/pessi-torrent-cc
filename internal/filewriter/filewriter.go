package filewriter

import (
	"PessiTorrent/internal/logger"
	"PessiTorrent/internal/utils"
	"os"
	"sync"
)

const (
	Permissions     = 0666
	Flags           = os.O_WRONLY | os.O_CREATE

	WorkerPoolSize = 10
)

type FileWriter struct {
	file        *os.File
	fileName    string
	chunkSize   uint64
	chunksQueue chan Chunk
	onWrite     func(index uint16)
	stopChannel chan struct{}
	workerWg    sync.WaitGroup
}

type Chunk struct {
	index uint16
	data  []uint8
}

func NewFileWriter(fileName string, fileSize uint64, onWrite func(index uint16), downloadsFolder string) (*FileWriter, error) {
	path := downloadsFolder + "/" + fileName

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
		onWrite:     onWrite,
		stopChannel: make(chan struct{}),
	}, nil
}

func (fileWriter *FileWriter) EnqueueChunkToWrite(index uint16, data []uint8) {
	fileWriter.chunksQueue <- Chunk{index, data}
}

func (fileWriter *FileWriter) Start() {
	for i := 0; i < WorkerPoolSize; i++ {
		fileWriter.workerWg.Add(1)
		go fileWriter.workerPool()
	}

	<-fileWriter.stopChannel
	fileWriter.workerWg.Wait()
}

func (fileWriter *FileWriter) workerPool() {
	for {
		chunk, ok := <-fileWriter.chunksQueue
		if !ok {
			fileWriter.workerWg.Done()
			return
		}

		fileWriter.writeChunk(chunk)
	}
}

func (fileWriter *FileWriter) writeChunk(chunk Chunk) {
	_, err := fileWriter.file.WriteAt(chunk.data, int64(chunk.index)*int64(fileWriter.chunkSize))
	if err != nil {
		logger.Error("Error writing chunk to file: %v", err)
	}

	//logger.Info("Chunk of index %d written to file %s", chunk.index, fw.fileName)
	fileWriter.onWrite(chunk.index)
}

func (fileWriter *FileWriter) Stop() {
	err := fileWriter.file.Close()
	if err != nil {
		logger.Error("Error closing file: %v", err)
	}

	fileWriter.stopChannel <- struct{}{}
	close(fileWriter.chunksQueue)
	close(fileWriter.stopChannel)
}
