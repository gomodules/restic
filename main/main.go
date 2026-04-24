package main

import (
	"fmt"
	"os"
	"os/exec"
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

	randomFilePath := "random_5G.bin"

	//dd if=/dev/urandom of=random_5G.bin bs=1M count=5124 status=progress
	//writeRandomFile(randomFilePath)

	pipeCommand := restic.Command{
		Name: "cat",
		Args: []any{randomFilePath},
	}
	backupOpt := restic.BackupOptions{
		StdinPipeCommands: []restic.Command{pipeCommand},
		StdinFileName:     randomFilePath,
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
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

	go func() {
		var since int = 0
		for {
			length, out, err := w.LeafOutput("anisur", since)
			if err != nil {
				fmt.Println("Failed to get leaf output:", err)
			} else {
				if len(out) > 0 {
					//fmt.Println(out)
				}
			}
			if length > since {
				fmt.Println(out)
			}
			since = length
			time.Sleep(1 * time.Second)
		}
	}()

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

	cmd := exec.Command(
		"dd",
		"if=/dev/zero",
		"of="+file.Name(),
		"bs=1M",
		"count=5120", // 5 GiB
		"status=progress",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = nil
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
	fmt.Println("Random File Write Done:", randomFilePath)
}
