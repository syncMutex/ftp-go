package httpServer

import (
	"encoding/json"
	"fmt"
	"ftp/common"
	serverUtils "ftp/server/server-utils"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

type ErrStruct struct {
	Err    bool   `json:"err"`
	ErrMsg string `json:"errMsg"`
}

func curWorkingDir(w http.ResponseWriter, r *http.Request) {
	if dirPath, err := serverUtils.GetAbsPath("./"); err == nil {
		w.Write([]byte(filepath.ToSlash(dirPath)))
	}
}

func ls(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	reqBody, _ := ioutil.ReadAll(r.Body)

	var path struct {
		Path string `json:"path"`
	}

	err := json.Unmarshal(reqBody, &path)

	var filesResponse struct {
		ErrStruct
		Files []common.FileStruct `json:"files"`
	}

	if err != nil {
		filesResponse.Err = true
		filesResponse.ErrMsg = err.Error()
	} else {
		filesResponse.Files, err = serverUtils.GetFileList(path.Path)
		if err != nil {
			filesResponse.Err = true
			filesResponse.ErrMsg = err.Error()
		}
	}
	json.NewEncoder(w).Encode(filesResponse)
}

func pathExists(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	reqBody, _ := ioutil.ReadAll(r.Body)

	var path struct {
		Path string `json:"path"`
	}

	err := json.Unmarshal(reqBody, &path)

	var filesResponse struct {
		ErrStruct
		PathExists bool `json:"pathExists"`
	}

	if err != nil {
		filesResponse.Err = true
		filesResponse.ErrMsg = err.Error()
	} else {
		filesResponse.PathExists = common.PathExists(path.Path)
	}
	json.NewEncoder(w).Encode(filesResponse)
}

func getFiles(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	fileName := serverUtils.GetFileName(string(body))

	filePath, _ := filepath.Abs(fileName)

	os.Mkdir("./.tmp", os.ModePerm)
	zipPath := "./.tmp/" + fileName + ".zip"
	common.ZipSource([]string{filePath}, zipPath, nil)

	http.ServeFile(w, r, zipPath)

	os.RemoveAll("./.tmp/")
}

func getMultipleFiles(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)

	var paths []string

	err := json.Unmarshal(body, &paths)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	os.Mkdir("./.tmp", os.ModePerm)
	fileName := "download"
	zipPath := "./.tmp/" + fileName + ".zip"
	if err := common.ZipSource(paths, zipPath, nil); err != nil {
		fmt.Println(err.Error())
		return
	}
	http.ServeFile(w, r, zipPath)
	os.RemoveAll("./.tmp/")
}

func printNetworks(port string) {
	fmt.Println("running on:")
	fmt.Println("    http://localhost" + port)
	if ipv4 := common.GetIPv4Str(); ipv4 != common.LOCAL_HOST {
		fmt.Println("    http://" + ipv4 + port)
	}
}

func StartHttpServer(PORT string) {
	r := mux.NewRouter()

	r.HandleFunc("/pwd", curWorkingDir)
	r.HandleFunc("/ls", ls).Methods("POST")
	r.HandleFunc("/path-exists", pathExists).Methods("POST")
	r.HandleFunc("/get", getFiles).Methods("POST")
	r.HandleFunc("/get-multiple", getMultipleFiles).Methods("POST")
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./server/http-server/public/"))))
	printNetworks(":" + PORT)

	if err := http.ListenAndServe(":"+PORT, r); err != nil {
		log.Fatal("Error Starting the HTTP Server :", err)
		return
	}
}
