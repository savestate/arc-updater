// Intended to be placed inside the Guild Wars 2 install directory.

package main

import (
	"bufio"
	"crypto/md5"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	MD5FileName  string = "d3d11.dll.md5sum"
	dllFileName  string = "d3d11.dll"
	transportTLS        = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	requestClient = http.Client{
		Timeout:   5 * time.Second,
		Transport: transportTLS,
	}
)

func main() {
	targetSlice := []string{
		"https://www.deltaconnected.com/arcdps/",
		"https://www.deltaconnected.com/arcdps/x64/",
		"https://www.deltaconnected.com/arcdps/x64/d3d11.dll",
		"https://www.deltaconnected.com/arcdps/x64/d3d11.dll.md5sum",
	}
	if request(targetSlice[0]) { // If the site is responding
		if !checkDllExists() {
			fmt.Println("Arcdps not found - Installing")
			site := getSiteMD5(targetSlice[3])
			local := getNewVersion(targetSlice[2])
			if site == local {
				fmt.Println("Checksum OK - Installing Arcdps.")
				err := os.Rename("./d3d11.dll", "../d3d11.dll")
				if err != nil {
					fmt.Printf("%v", err)
				}
				err = os.Remove(MD5FileName)
				if err != nil {
					fmt.Printf("%v", err)
				}
				return
			} else {
				fmt.Println("Checksums do not match - Arcdps will not be Installed.")
				err := os.Remove(dllFileName)
				if err != nil {
					fmt.Printf("%v", err)
				}
				err = os.Remove(MD5FileName)
				if err != nil {
					fmt.Printf("%v", err)
				}
				return
			}
		}
		site := getSiteMD5(targetSlice[3])
		local := getLocalMD5()
		if site != local { //if the sums are different for local version and site version
			new := getNewVersion(targetSlice[2])
			if site == new {
				fmt.Println("Checksum OK - Updating Arcdps.")
				err := os.Rename("./d3d11.dll", "../d3d11.dll")
				if err != nil {
					fmt.Printf("%v", err)
				}
				err = os.Remove(MD5FileName)
				if err != nil {
					fmt.Printf("%v", err)
				}
				return
			} else {
				fmt.Println("Checksums do not match - Arcdps will not be updated.")
				err := os.Remove(dllFileName)
				if err != nil {
					fmt.Printf("%v", err)
				}
				err = os.Remove(MD5FileName)
				if err != nil {
					fmt.Printf("%v", err)
				}
				return
			}
		} else {
			fmt.Println("Your Arcdps is already the latest version. Exiting.")
			err := os.Remove(MD5FileName)
			if err != nil {
				fmt.Printf("%v", err)
			}
		}
	} else {
		fmt.Println("Bad Response from https://www.deltaconnected.com/arcdps/ - Exiting.")
	}
	fmt.Println("Finished - Press any key to exit.")
	fmt.Scanln()
}

func request(target string) bool {
	response, err := requestClient.Get(target)
	if err != nil {
		fmt.Println(err)
		return false
	}
	if response.StatusCode == 200 {
		return true
	} else {
		return false
	}
}

func getSiteMD5(md5URL string) string {
	sumfile, err := os.Create(MD5FileName)
	if err != nil {
		panic(err)
	}
	defer sumfile.Close()

	resp, err := requestClient.Get(md5URL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(sumfile, resp.Body)
	if err != nil {
		panic(err)
	}

	oFile, err := os.Open(sumfile.Name())
	if err != nil {
		panic(err)
	}
	defer oFile.Close()

	oFileScanner := bufio.NewScanner(oFile)
	oFileScanner.Scan()
	split := strings.Fields(oFileScanner.Text())
	return split[0]
}

func getLocalMD5() string {
	localdll := "../d3d11.dll"
	sumfile, err := os.Open(localdll)
	if err != nil {
		panic(err)
	}
	defer sumfile.Close()

	hash := md5.New()

	_, err = io.Copy(hash, sumfile)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func getNewVersion(dllURL string) string {
	file, err := os.Create(dllFileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	resp, err := requestClient.Get(dllURL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		panic(err)
	}

	//verify downloaded sum matches the posted sum
	hash := md5.New()
	nFile, err := os.Open(dllFileName)
	if err != nil {
		panic(err)
	}
	defer nFile.Close()
	_, err = io.Copy(hash, nFile)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func checkDllExists() bool {
	if _, err := os.Stat("../d3d11.dll"); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Println("d3d11.dll Does Not Exist.")
			return false
		} else {
			panic(err)
		}
	}
	fmt.Println("d3d11.dll Exists.")
	return true
}
