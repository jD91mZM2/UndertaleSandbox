package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/legolord208/stdutil"
)

var regexFile0 = regexp.MustCompile(`file0( \([0-9]+\))?`)
var regexINI = regexp.MustCompile(`undertale( \([0-9]+\))?.ini`)

const roomLine = 548

const configFile = ".undertalesandbox"

type configStruct struct {
	DownloadsDir       string
	UndertaleConfigDir string
	UndertaleBinaryDir string
}

var config configStruct

func main() {
	var flagFile0 string
	var flagINI string
	var flagRestart bool

	flag.BoolVar(&flagRestart, "r", false, "Ask to restart undertale if closed?")
	flag.StringVar(&flagFile0, "file0", "", "Manually specify 'file0' location (default scans in downloads folder)")
	flag.StringVar(&flagINI, "ini", "", "Manually specify 'undertale.ini' location (default scans in downloads folder)")
	flag.Parse()

	fmt.Println("Reading config...")

	// Compatibility reasons.
	_, err := os.Stat(".ui_config")
	if err != nil {
		if os.IsNotExist(err) {
			stdutil.PrintErr("Couldn't stat .ui_config", err)
		}
	} else {
		// Old config name.
		// Should be renamed
		err := os.Rename(".ui_config", configFile)
		if err != nil {
			stdutil.PrintErr("Attempt to rename old config file failed.", err)
		}
	}

	file, err := os.Open(configFile)
	if err != nil {
		if !os.IsNotExist(err) {
			stdutil.PrintErr("Could not read file", err)
			return
		}
		fmt.Println("Does not exist. Creating...")

		current, err := user.Current()
		if err != nil {
			stdutil.PrintErr("Could not find current user", err)
			return
		}

		config = configStruct{
			DownloadsDir:       filepath.Join(current.HomeDir, "Downloads"),
			UndertaleConfigDir: defaultConfigDir(current),
			UndertaleBinaryDir: filepath.Join(defaultSteamDir(current), "steamapps", "common", "Undertale"),
		}

		file, err = os.Open(configFile)
		if err != nil {
			stdutil.PrintErr("Could not generate default config", err)
		} else {
			encoder := json.NewEncoder(file)
			encoder.SetIndent("", "\t")
			err := encoder.Encode(config)
			file.Close()
			if err != nil {
				stdutil.PrintErr("Could not save default config", err)
			}
		}
	} else {
		err = json.NewDecoder(file).Decode(&config)
		file.Close()
		if err != nil {
			stdutil.PrintErr("Could not parse JSON", err)
			return
		}
	}

	var locFile0 string
	var locINI string

	dstLocFile0 := filepath.Join(config.UndertaleConfigDir, "file0")
	dstLocINI := filepath.Join(config.UndertaleConfigDir, "undertale.ini")

	unsetFile0 := flagFile0 == ""
	unsetINI := flagINI == ""

	if unsetFile0 || unsetINI {
		fmt.Println("Finding files...")

		files, err := ioutil.ReadDir(config.DownloadsDir)
		if err != nil {
			if os.IsNotExist(err) {
				stdutil.PrintErr("Downloads directory does not exist", nil)
			} else {
				stdutil.PrintErr("Could not scan directory", err)
			}
			checkConf()
			return
		}

		var file0 os.FileInfo
		var ini os.FileInfo
		for _, file := range files {
			if unsetFile0 {
				if regexFile0.MatchString(file.Name()) &&
					(file0 == nil || file.ModTime().After(file0.ModTime())) {
					file0 = file
				}
			}
			if unsetINI {
				if regexINI.MatchString(file.Name()) &&
					(ini == nil || file.ModTime().After(ini.ModTime())) {
					ini = file
				}
			}
		}

		if !unsetFile0 {
			locFile0 = flagFile0
		} else if file0 == nil {
			locFile0 = dstLocFile0 + ".back" // Because it'll be moved later on
		} else {
			locFile0 = filepath.Join(config.DownloadsDir, file0.Name())
		}

		if !unsetINI {
			locINI = flagINI
		} else if ini == nil {
			locINI = dstLocINI + ".back" // Because it'll be moved later on
		} else {
			locINI = filepath.Join(config.DownloadsDir, ini.Name())
		}
	} else {
		locFile0 = flagFile0
		locINI = flagINI
	}

	fmt.Println("file0: " + locFile0)
	fmt.Println("undertale.ini: ", locINI)

	fmt.Println("Checking for existing backups...")
	_, err = os.Stat(dstLocFile0 + ".back")
	if err == nil || !os.IsNotExist(err) {
		stdutil.PrintErr("Backup for file0 already exists. Cancelling!", nil)
		return
	}

	_, err = os.Stat(dstLocINI + ".back")
	if err == nil || !os.IsNotExist(err) {
		stdutil.PrintErr("Backup for file0 already exists. Cancelling!", nil)
		return
	}

	fmt.Println("Backing up...")
	err = os.Rename(dstLocFile0, dstLocFile0+".back")
	if err != nil {
		stdutil.PrintErr("Could not backup file0", err)
		checkConf()
		return
	}

	defer func() {
		fmt.Println("Restoring file0...")
		err := os.Rename(dstLocFile0+".back", dstLocFile0)
		if err != nil {
			stdutil.PrintErr("Could not restore file0 backup", err)
			return
		}
	}()

	err = os.Rename(dstLocINI, dstLocINI+".back")
	if err != nil {
		stdutil.PrintErr("Could not backup undertale.ini", err)
		return
	}

	defer func() {
		fmt.Println("Restoring undertale.ini...")
		err := os.Rename(dstLocINI+".back", dstLocINI)
		if err != nil {
			stdutil.PrintErr("Could not restore ini backup", err)
			return
		}
	}()

	fmt.Println("Cloning...")
	success, room := cloneFile0(locFile0, dstLocFile0)
	if !success {
		return
	}

	success = cloneINI(locINI, dstLocINI, room)
	if !success {
		return
	}

	fmt.Println("Starting!")

loop:
	for {
		cmd := command()
		cmd.Dir = config.UndertaleBinaryDir
		err = cmd.Run()
		if err != nil {
			stdutil.PrintErr("Couldn't run undertale", err)
		}

		if flagRestart {
			for {
				fmt.Print("Undertale closed. Restart? [y/n] ")
				opt := stdutil.MustScanLower()

				if opt == "y" {
					continue loop
				} else if opt == "n" {
					break
				}

				fmt.Println("Not a valid option. Must be either 'y' or 'n'.")
			}
		}
		break
	}
}

func cloneFile0(src, dst string) (success bool, room string) {
	srcFile0, err := os.Open(src)
	if err != nil {
		if os.IsNotExist(err) {
			stdutil.PrintErr("Hey, that file0 does not exist!", nil)
		} else {
			stdutil.PrintErr("Error reading file0", err)
		}
		return
	}
	defer srcFile0.Close()

	dstFile0, err := os.Create(dst)
	if err != nil {
		stdutil.PrintErr("Could not open file0 destination", nil)
		return
	}
	defer dstFile0.Close()

	scanner := bufio.NewScanner(srcFile0)
	i := 0
	for scanner.Scan() {
		text := scanner.Text()

		_, err := dstFile0.WriteString(text + "\r\n") // DOS file endings :(
		if err != nil {
			stdutil.PrintErr("Could not write to destination file0", err)
			return
		}

		i++
		if i == roomLine {
			room = strings.TrimSpace(text)
		}
	}

	err = scanner.Err()
	if err != nil {
		stdutil.PrintErr("Error scanning file0", err)
		return
	}

	success = true
	return
}

func cloneINI(src, dst, room string) (success bool) {
	srcINI, err := os.Open(src)
	if err != nil {
		stdutil.PrintErr("Error reading undertale.ini", err)
		checkConf()
		return
	}
	defer srcINI.Close()

	dstINI, err := os.Create(dst)
	if err != nil {
		stdutil.PrintErr("Error opening destination undertale.ini", err)
		return
	}
	defer dstINI.Close()

	scanner := bufio.NewScanner(srcINI)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "Room=") {
			fmt.Println("Making sure annoying dog doesn't appear...")
			text = text[:5] + "\"" + room + "\""
		}

		_, err := dstINI.WriteString(text + "\r\n") // DOS file endings :(
		if err != nil {
			stdutil.PrintErr("Could not write to destination file0", err)
			return
		}
	}

	err = scanner.Err()
	if err != nil {
		stdutil.PrintErr("Error scanning undertale.ini", err)
		return
	}

	success = true
	return
}

func checkConf() {
	stdutil.PrintErr("Please check all values in "+configFile, nil)
}
