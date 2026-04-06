package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"gomodules.xyz/restic"
	core "k8s.io/api/core/v1"
	storage "kmodules.xyz/objectstore-api/api/v1"
)

func main() {

	storageSecret := &core.Secret{
		Data: map[string][]byte{
			"RESTIC_PASSWORD": []byte("123"),
		},
	}

	setupOpt := restic.SetupOptions{
		Backends: []*restic.Backend{
			{
				StorageConfig: &restic.StorageConfig{
					Provider: storage.ProviderS3,
					Endpoint: "https://s3.amazonaws.com",
					Region:   "us-east-2",
					Bucket:   "kubestash",
					Prefix:   "leaf-stdout",
				},
				EncryptionSecret: storageSecret,
				Repository:       "anisur",
			},
		},
	}

	w, err := restic.NewResticWrapper(&setupOpt)
	if err != nil {
		fmt.Println("Failed to create ResticWrapper:", err)
		return
	}
	yes := w.RepositoryAlreadyExist("anisur")
	fmt.Println("RepositoryAlreadyExist:", yes)
	if !yes {
		err = w.InitializeRepository("anisur")
		if err != nil {
			fmt.Println("Failed to initialize repository:", err)
			return
		}
	}

	randomFilePath := "random_file.bin"

	writeRandomFile(randomFilePath)

	pipeCommand := restic.Command{
		Name: "cat",
		Args: []any{randomFilePath},
	}
	backupOpt := restic.BackupOptions{
		StdinPipeCommands: []restic.Command{pipeCommand},
		StdinFileName:     "random_file.bin",
	}
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		backupOut, err := w.RunBackup(backupOpt)
		if err != nil {
			fmt.Println("Failed to backup:", err)
			return
		}
		fmt.Println("backup output:")
		for _, out := range backupOut {
			fmt.Println(out)
		}
		wg.Done()
	}()

	for {
		out, err := w.LeafOutput("anisur")
		if err != nil {
			fmt.Println("Failed to get leaf output:", err)
		} else {
			if len(out) > 0 {
				fmt.Println(out)
			}
		}
		time.Sleep(3 * time.Second)

	}

	wg.Wait()

}

func writeRandomFile(randomFilePath string) {
	var notExist bool
	// Open file for writing
	_, err := os.Stat(randomFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			notExist = true
			// File does not exist, proceed to create it
		} else {
			panic(err)

		}
	}
	if !notExist {
		return
	}

	file, err := os.Create(randomFilePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 2 GB in bytes
	const fileSize = 1 * 1024 * 1024 * 1024
	// Chunk size: 1 MB
	const chunkSize = 1 * 1024 * 1024

	data := make([]byte, chunkSize)
	// Fill chunk with some pattern (deterministic)
	for i := range data {
		data[i] = 'A' // or any deterministic pattern
	}

	written := int64(0)
	for written < fileSize {
		if remaining := fileSize - written; remaining < chunkSize {
			_, err = file.Write(data[:remaining])
			written += remaining
		} else {
			_, err = file.Write(data)
			written += chunkSize
		}
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Random File Write Done:", randomFilePath)
}
