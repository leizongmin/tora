package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func cmdPut(args []string, cmd *flag.FlagSet, options baseOptions) {
	cmd.Parse(args)

	remotePath := cmd.Arg(0)
	if len(remotePath) < 1 {
		fmt.Println("Missing first argument <remotePath>")
	}
	remotePath = formatRemotePath(remotePath)
	fmt.Println("Remote Path:", remotePath)

	localPath := cmd.Arg(1)
	if len(localPath) < 1 {
		fmt.Println("Missing second argument <localPath>")
		os.Exit(1)
	}
	localPath = formatLocalPath(localPath)
	fmt.Println("Local Path: ", localPath)

	client := NewClient(options.server, options.token)

	info, err := os.Stat(localPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if info.IsDir() {
		fmt.Println("File Type:   Dir")
		err = uploadDir(client, remotePath, localPath)
	} else {
		fmt.Println("File Type:   file")
		err = uploadFile(client, remotePath, localPath)
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func uploadDir(client *Client, remotePath string, localPath string) error {
	files, err := ioutil.ReadDir(localPath)
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() {
			err := uploadDir(client, remotePath+"/"+f.Name(), filepath.Join(localPath, f.Name()))
			if err != nil {
				return err
			}
		} else {
			err := uploadFile(client, remotePath+"/"+f.Name(), filepath.Join(localPath, f.Name()))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func uploadFile(client *Client, remotePath string, localPath string) error {
	fmt.Printf("Upload: [%s] %s\n", remotePath, localPath)
	md5, err := getFileMd5(localPath)
	if err != nil {
		return err
	}
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	req, err := client.Put("file", remotePath, file)
	if err != nil {
		return err
	}
	req.Header.Set("x-content-md5", md5)
	_, data, err := client.ResponseJson(req)
	if err != nil {
		return err
	}
	if data.Get("ok").ToBool() {
		fmt.Printf("  - Success: md5=%s checked=%s\n", md5, data.Get("data", "checkedMd5").ToString())
		return nil
	}
	return fmt.Errorf("upload failed: %s", data.Get("error"))
}

func getFileMd5(filePath string) (string, error) {
	var returnMD5String string
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}
	hashInBytes := hash.Sum(nil)[:16]
	returnMD5String = hex.EncodeToString(hashInBytes)
	return returnMD5String, nil
}

func formatRemotePath(remotePath string) string {
	if remotePath[0:1] != "/" {
		remotePath = "/" + remotePath
	}
	if remotePath[len(remotePath)-1:] == "/" {
		remotePath = remotePath[0 : len(remotePath)-1]
	}
	return remotePath
}

func formatLocalPath(localPath string) string {
	localPath, err := filepath.Abs(localPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return localPath
}
