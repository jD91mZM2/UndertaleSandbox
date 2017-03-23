package main;

import (
	"fmt"
	"github.com/legolord208/stdutil"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"encoding/json"
	"regexp"
	"bufio"
	"flag"
	"strings"
)

var R_FILE0 = regexp.MustCompile(`file0( \([0-9]+\))?`);
var R_INI = regexp.MustCompile(`undertale( \([0-9]+\))?.ini`);

const FILE0_ROOM_LINE = 548;

const CONFIG_NAME = ".ui_config";
type configStruct struct{
	DownloadsDir string
	UndertaleConfigDir string
	UndertaleBinaryDir string
}
var config configStruct;

func main(){
	var restart bool;

	flag.BoolVar(&restart, "r", false, "Ask to restart undertale if closed?");
	flag.Parse();

	fmt.Println("Reading config...");

	contents, err := ioutil.ReadFile(CONFIG_NAME);
	if(err != nil){
		if(!os.IsNotExist(err)){
			stdutil.PrintErr("Could not read file", err);
			return;
		}
		fmt.Println("Does not exist. Creating...");

		current, err := user.Current();
		if(err != nil){
			stdutil.PrintErr("Could not find current user", err);
			return;
		}

		config = configStruct{
			DownloadsDir: filepath.Join(current.HomeDir, "Downloads"),
			UndertaleConfigDir: defaultConfigDir(current),
			UndertaleBinaryDir: filepath.Join(defaultSteamDir(current), "steamapps", "common", "Undertale"),
		};

		contents, err = json.MarshalIndent(config, "", "\t");
		if(err != nil){
			stdutil.PrintErr("Could not generate default config", err);
		} else {
			err = ioutil.WriteFile(CONFIG_NAME, contents, 0666);
			if(err != nil){
				stdutil.PrintErr("Could not save default config", err);
			}
		}
	} else {
		err = json.Unmarshal(contents, &config);
		if(err != nil){
			stdutil.PrintErr("Could not parse JSON", err);
			return;
		}
	}

	fmt.Println("Finding files...");

	files, err := ioutil.ReadDir(config.DownloadsDir);
	if(err != nil){
		if(os.IsNotExist(err)){
			stdutil.PrintErr("Downloads directory does not exist", nil);
		} else {
			stdutil.PrintErr("Could not scan directory", err);
		}
		checkConf();
		return;
	}

	var file0 os.FileInfo;
	var ini os.FileInfo;
	for _, file := range files{
		if(R_FILE0.MatchString(file.Name()) &&
				(file0 == nil || file.ModTime().After(file0.ModTime()))){
			file0 = file;
		}
		if(R_INI.MatchString(file.Name()) &&
				(ini == nil || file.ModTime().After(ini.ModTime()))){
			ini = file;
		}
	}

	var loc_file0 string;
	var loc_ini string;
	dst_loc_file0 := filepath.Join(config.UndertaleConfigDir, "file0");
	dst_loc_ini := filepath.Join(config.UndertaleConfigDir, "undertale.ini");

	if(file0 == nil){
		stdutil.PrintErr("Could not find file0 in downloads directory!", nil);
		fmt.Print("file0 path: ");
		loc_file0 = stdutil.MustScanTrim();
	}
	loc_file0 = filepath.Join(config.DownloadsDir, file0.Name());

	if(ini == nil){
		loc_ini = dst_loc_ini + ".back"; // Because it'll be moved later on
	} else {
		loc_ini = filepath.Join(config.DownloadsDir, ini.Name());
	}

	fmt.Println("file0: " + loc_file0);
	fmt.Println("undertale.ini: ", loc_ini);

	fmt.Println("Checking for existing backups...");
	_, err = os.Stat(dst_loc_file0 + ".back");
	if(err == nil || !os.IsNotExist(err)){
		stdutil.PrintErr("Backup for file0 already exists. Cancelling!", nil);
		return;
	}

	_, err = os.Stat(dst_loc_ini + ".back");
	if(err == nil || !os.IsNotExist(err)){
		stdutil.PrintErr("Backup for file0 already exists. Cancelling!", nil);
		return;
	}

	fmt.Println("Backing up...");
	err = os.Rename(dst_loc_file0, dst_loc_file0 + ".back");
	if(err != nil){
		stdutil.PrintErr("Could not backup file0", err);
		checkConf();
		return;
	}

	defer func(){
		fmt.Println("Restoring file0...");
		err := os.Rename(dst_loc_file0 + ".back", dst_loc_file0);
		if(err != nil){
			stdutil.PrintErr("Could not restore file0 backup", err);
			return;
		}
	}();

	err = os.Rename(dst_loc_ini, dst_loc_ini + ".back");
	if(err != nil){
		stdutil.PrintErr("Could not backup undertale.ini", err);
		return;
	}

	defer func(){
		fmt.Println("Restoring undertale.ini...");
		err := os.Rename(dst_loc_ini + ".back", dst_loc_ini);
		if(err != nil){
			stdutil.PrintErr("Could not restore ini backup", err);
			return;
		}
	}();

	fmt.Println("Cloning...");
	success, room := cloneFile0(loc_file0, dst_loc_file0);
	if(!success){
		return;
	}

	success = cloneINI(loc_ini, dst_loc_ini, room);
	if(!success){
		return;
	}

	fmt.Println("Starting!");

	loop:
	for{
		cmd := command();
		cmd.Dir = config.UndertaleBinaryDir;
		err = cmd.Run();
		if(err != nil){
			stdutil.PrintErr("Couldn't run undertale", err);
			return;
		}

		if(restart){
			for{
				fmt.Print("Undertale closed. Restart? [y/n] ");
				opt := stdutil.MustScanLower();

				if(opt == "y"){
					continue loop;
				} else if(opt == "n"){
					break;
				}

				fmt.Println("Not a valid option. Must be either 'y' or 'n'.");
			}
		}
		break;
	}
}

func cloneFile0(src, dst string) (success bool, room string){
	src_file0, err := os.Open(src);
	if(err != nil){
		if(os.IsNotExist(err)){
			stdutil.PrintErr("Hey, that file0 does not exist!", nil);
		} else {
			stdutil.PrintErr("Error reading file0", err);
		}
		return;
	}
	defer src_file0.Close();

	dst_file0, err := os.Create(dst);
	if(err != nil){
		stdutil.PrintErr("Could not open file0 destination", nil);
		return;
	}
	defer dst_file0.Close();

	scanner := bufio.NewScanner(src_file0);
	i := 0;
	for scanner.Scan(){
		text := scanner.Text();

		_, err := dst_file0.WriteString(text + "\r\n"); // DOS file endings :(
		if(err != nil){
			stdutil.PrintErr("Could not write to destination file0", err);
			return;
		}

		i++;
		if(i == FILE0_ROOM_LINE){
			room = strings.TrimSpace(text);
		}
	}

	err = scanner.Err();
	if(err != nil){
		stdutil.PrintErr("Error scanning file0", err);
		return;
	}

	success = true;
	return;
}

func cloneINI(src, dst, room string) (success bool){
	src_ini, err := os.Open(src);
	if(err != nil){
		stdutil.PrintErr("Error reading undertale.ini", err);
		checkConf();
		return;
	}
	defer src_ini.Close();

	dst_ini, err := os.Create(dst);
	if(err != nil){
		stdutil.PrintErr("Error opening destination undertale.ini", err);
		return;
	}
	defer dst_ini.Close();

	scanner := bufio.NewScanner(src_ini);
	for scanner.Scan(){
		text := scanner.Text();
		if(strings.HasPrefix(text, "Room=")){
			fmt.Println("Making sure annoying dog doesn't appear...");
			text = text[:5] + "\"" + room + "\"";
		}

		_, err := dst_ini.WriteString(text + "\r\n"); // DOS file endings :(
		if(err != nil){
			stdutil.PrintErr("Could not write to destination file0", err);
			return;
		}
	}

	err = scanner.Err();
	if(err != nil){
		stdutil.PrintErr("Error scanning undertale.ini", err);
		return;
	}

	success = true;
	return;
}

func checkConf(){
	stdutil.PrintErr("Please check all values in " + CONFIG_NAME, nil);
}
