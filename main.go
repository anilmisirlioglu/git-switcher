// Copyright 2021 Kaan Karakaya
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/mitchellh/go-homedir"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func main() {
	confPath, err := homedir.Expand("~/.config/gitconfigs")
	if err != nil {
		log.Panic(err)
	}

	log.SetFlags(log.Lshortfile)
	// hash: filename
	configs := make(map[string]string)

	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		// Give permission for only current user
		err = os.Mkdir(confPath, os.ModeDir|0700)
		if err != nil {
			log.Println(err)
		}
	}

	// List ~/.gitconfigs folder
	err = filepath.WalkDir(confPath+"/", func(path string, d fs.DirEntry, e error) error {
		if d.IsDir() {
			return nil
		}

		configs[hash(path)] = filepath.Base(path)

		return nil
	})
	if err != nil {
		log.Println(err)
	}

	// Check current gitconfig is exist in configs
	gitConfig, _ := homedir.Expand("~/.gitconfig")

	// If gitconfig file is not exist create empty file
	if _, err := os.Stat(gitConfig); os.IsNotExist(err) {
		write(gitConfig, []byte(""))
	}
	gitConfigHash := hash(gitConfig)
	if _, ok := configs[gitConfigHash]; !ok {
		err := os.Link(gitConfig, confPath+"/Old Config")
		if err != nil {
			log.Panic(err)
		}
	}

	//	log.Println(configs)
	newConfig := ""
	if len(os.Args) > 1 {
		action := os.Args[1]
		switch action {
		case "create":
			prom := promptui.Prompt{
				Label: "Profile name",
			}

			result, err := prom.Run()
			if err != nil {
				log.Panic(err)
			}

			// File is not exist write
			if _, err := os.Stat(confPath + "/" + result); os.IsNotExist(err) {
				write(confPath+"/"+result, []byte("[user]\n\tname = "+result))
			} else {
				color.HiRed("Profile is already exist")
			}
		case "delete":
			// List git configs
			var profiles []string
			var pos int
			i := 0
			for hash, val := range configs {
				// Find current config index
				if hash == gitConfigHash {
					pos = i
				}
				profiles = append(profiles, val)
				i++
			}

			prompt := promptui.Select{
				Label: "Select Git Config (Current: " + configs[gitConfigHash] + ")",
				Items: profiles,
				// Change cursor to current config
				CursorPos:    pos,
				HideSelected: true,
			}

			_, result, err := prompt.Run()
			if err != nil {
				log.Panic(err)
			}

			prom := promptui.Prompt{
				Label: "Are you sure ? Y/N",
			}
			asd, err := prom.Run()
			if err != nil {
				log.Panic(err)
			}
			if asd == "y" || asd == "Y" {
				err = os.Remove(confPath + "/" + result)
				if err != nil {
					log.Panic(err)
				}
				color.HiBlue("Profile deleted %q", result)
			} else {
				color.HiBlue("Profile not deleted %q", result)
			}
		case "rename":
			prom := promptui.Prompt{
				Label: "Profile name",
			}

			result, err := prom.Run()
			if err != nil {
				log.Panic(err)
			}

			promD := promptui.Prompt{
				Label: "Desired Profile name",
			}

			resultD, err := promD.Run()
			if err != nil {
				log.Panic(err)
			}

			err = os.Rename(confPath+"/"+result, confPath+"/"+resultD)
			if err != nil {
				log.Panic(err)
			}

		}
	} else if len(configs) >= 1 {
		// List git configs
		var profiles []string
		var pos int
		i := 0
		for hash, val := range configs {
			// Find current config index
			if hash == gitConfigHash {
				pos = i
			}
			profiles = append(profiles, val)
			i++
		}

		prompt := promptui.Select{
			Label: "Select Git Config (Current: " + configs[gitConfigHash] + ")",
			Items: profiles,
			// Change cursor to current config
			CursorPos:    pos,
			HideSelected: true,
		}

		_, result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}
		newConfig = result
		// Remove file for link new one
		err = os.Remove(gitConfig)
		if err != nil {
			log.Panic(err)
		}

		// Symbolic link to "~/.gitconfig"
		err = os.Symlink(confPath+"/"+newConfig, gitConfig)
		if err != nil {
			log.Panic(err)
		}
		color.HiBlue("Switched to profile %q", newConfig)
	}
}

func hash(path string) string {
	f, _ := os.Open(path)
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func write(file string, data []byte) {
	f, e := os.Create(file)
	if e != nil {
		log.Panic(e)
	}

	defer f.Close()
	_, err := f.Write(data)
	if err != nil {
		log.Panic(err)
	}
}
