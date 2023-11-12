package main

import (
	"PessiTorrent/internal/packets"
	"fmt"
  "os"
  "bufio"
)

// request <file>
func (n *Node) requestFile(args []string) error {
	filename := args[0]

	var packet packets.RequestFilePacket
	packet.Create(filename)
	err := n.conn.WritePacket(packet)
	if err != nil {
		return err
	}

	// TODO: handle file not found packet
  fileHashesPacket, _, err := n.conn.ReadPacket()

  pf := fileHashesPacket.(*packets.PublishFilePacket)
  n.AddDownloadFile(pf.FileName, pf.FileHash, pf.ChunkHashes)

	// Read response packets
	responsePacket, _, err := n.conn.ReadPacket()
	fmt.Printf("node %v, %v has the file\n", responsePacket.(*packets.AnswerNodesPacket).NodeIPAddr, responsePacket.(*packets.AnswerNodesPacket).UDPPort)
	if err != nil {
		return err
	}

	// Read response packets
	for {
		if responsePacket.(*packets.AnswerNodesPacket).SequenceNumber == 0 {
			break
		}

		responsePacket, _, err = n.conn.ReadPacket()
		fmt.Printf("node %v, %v has the file\n", responsePacket.(*packets.AnswerNodesPacket).NodeIPAddr, responsePacket.(*packets.AnswerNodesPacket).UDPPort)
		if err != nil {
			return err
		}
	}

	return nil
}

// publish <file path>
func (n *Node) publishFile(args []string) error {
	filePath := args[0]
  f, err := n.AddSharedFile(filePath)
  if err != nil {
    return err
  }

	var packet packets.PublishFilePacket
	packet.Create(f.filename, f.fileHash, f.chunkHashes)
	err = n.conn.WritePacket(packet)
	if err != nil {
		return err
	}

	return nil
}

// load <file path>
func (n *Node) loadSharedFiles(args []string) error {
  filePath := args[0]

  file, err := os.Open(filePath)
  if err != nil {
    return err
  }
  defer file.Close()

  // read the file line by line
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    line := scanner.Text()
    // publish the file 
    n.publishFile([]string{line})
  }

  return nil
}

// check
func (n *Node) checkSharedFiles(args []string) error {
  n.files.Lock()
  defer n.files.Unlock()

  for _, file := range n.files.M {
    fmt.Println("----------------------------------------")
    fmt.Printf("File: %v\n", file.filename)
    fmt.Printf("Filepath: %v\n", file.filepath)
    fmt.Printf("File hash: %v\n", file.fileHash)
    fmt.Printf("Chunk hashes: %v\n", file.chunkHashes)
    fmt.Printf("Bitfield: %b\n", file.bitfield)
  }

  return nil
}

func (n *Node) removeSharedFile(args []string) error {
  filename := args[0]

  n.RemoveFile(filename)
	var packet packets.RemoveFilePacket
	packet.Create(filename)

  err := n.conn.WritePacket(packet)
  if err != nil {
    return err
  }

  return nil
}
