package main

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

/*
	Common data structure for both json, XML
 */
type root struct {
	XMLName xml.Name `xml:"map"`
	Version string   `xml:"version,attr" json:"map_version"`
	Node    node     `xml:"node" json:"root"`
}

type node struct {
	Text  string `xml:"TEXT,attr" json:"title"`
	Nodes []node `xml:"node" json:"children"`
}

/*
	Unzips the .mind file locally (temp folder) and converts the json to XML/mm format
 */
func json2xml(dotMindFile string, fmFilename string) {

	dir, err := ioutil.TempDir(".", "tmp")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	UnzipFiles(dotMindFile, dir)

	jsonFile, err1 := os.Open(dir+"/map.json")
	if err1 != nil {
		fmt.Println("Opening file error : ", err1)
		return
	}
	defer jsonFile.Close()

	jsonData, _ := ioutil.ReadAll(jsonFile)

	var m root

	err = json.Unmarshal([]byte(jsonData), &m)

	if err != nil {
		panic(err)
	}

	log.Println(reflect.TypeOf(m))
	log.Println(m.Version)
	log.Println(m.Node.Text)
	log.Println(m.Node.Nodes)

	m.Version = "1.0.1"

	var y = m

	//enc := xml.NewEncoder(os.Stdout)
	//enc.Indent("", "  ")
	//if err := enc.Encode(y); err != nil {
	//	fmt.Printf("error: %v\n", err)
	//}

	xmlContent, _ := xml.Marshal(y)
	err = ioutil.WriteFile(fmFilename, xmlContent, 0644)

}

func xml2json(fmFilename string, dotMindFilename string) {
	file, err1 := os.Open(fmFilename)
	if err1 != nil {
		fmt.Println("Opening file error : ", err1)
		return
	}
	defer file.Close()

	data, _ := ioutil.ReadAll(file)

	var m root

	err := xml.Unmarshal([]byte(data), &m)

	if err != nil {
		panic(err)
	}


	log.Println(reflect.TypeOf(m))
	log.Println(m.Version)
	log.Println(m.Node.Text)
	log.Println(m.Node.Nodes)

	m.Version = "2.6" // for .mind file ...

	var y = m

	//enc := json.NewEncoder(os.Stdout)
	//if err := enc.Encode(y); err != nil {
	//	fmt.Printf("error: %v\n", err)
	//}
	jsonContent, _ := json.Marshal(y)
	err = ioutil.WriteFile("map.json", jsonContent, 0644)

	ZipFiles(dotMindFilename, []string{"map.json" })

}

func UnzipFiles(zipFilename string, destFolder string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(zipFilename)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		defer rc.Close()

		fpath := filepath.Join(destFolder, f.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(destFolder)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: Path contains ZipSlip issues", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {

			os.MkdirAll(fpath, os.ModePerm)

		} else {

			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return filenames, err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, err
			}

			_, err = io.Copy(outFile, rc)

			outFile.Close()

			if err != nil {
				return filenames, err
			}

		}
	}
	return filenames, nil
}

func ZipFiles(zipFilename string, filesToZip []string) error {

	newZipFile, err := os.Create(zipFilename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	for _, file := range filesToZip {

		zipfile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer zipfile.Close()

		info, err := zipfile.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = file
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err = io.Copy(writer, zipfile); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	infilePtr := flag.String("in", "mandatory", "Name of the input freemind / dotMind file")
	outfilePtr := flag.String("out", "mandatory", "Name of the input freemind / dotMind file")
	modePtr := flag.Bool("j2m", true, "Mode: default .mind>.mm ; if false .mm > .mind (override with -j2m=false)")
	dbgPtr := flag.Bool("d", false, "Enable debug (by calling -d)")

	flag.Parse()

	/*
		Ensure input file exists AND
		Ensure output file does NOT exist (do not clobber)
	 */
	if strings.Compare(*infilePtr, "mandatory") == 0 {
		fmt.Println("Input filename is mandatory")
		return
	}

	if strings.Compare(*outfilePtr, "mandatory") == 0 {
		fmt.Println("Output filename is mandatory")
		return
	}

	if _, err := os.Stat(*infilePtr); os.IsNotExist(err) {
		fmt.Println(fmt.Sprintf("The input file '%s' does not exist", *infilePtr))
		return
	}

	if _, err := os.Stat(*outfilePtr); !os.IsNotExist(err) {
		fmt.Println(fmt.Sprintf("Output file '%s' already exists - will not overwrite", *outfilePtr))
		return
	}

	if *dbgPtr == true {
		log.SetFlags(1)
	} else {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	if *modePtr == true {
		fmt.Println("Converting from json to xml")
		json2xml(*infilePtr, *outfilePtr)
	} else {
		fmt.Println("Converting from xml to json")
		xml2json(*infilePtr, *outfilePtr)
	}

}
