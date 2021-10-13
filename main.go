package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// cutThis reads file per line and trims comment lines
func cutThis(filePath string) (string, error) {

	var blockStarted bool
	// Regex to match comments of type such as: "//comment", "// comment" and so on.
	regBasicComments := regexp.MustCompile(`([ \t]*\/\/.*)`)

	// Regex to match multiline comments of type such as: "/**/", "/* \n *\" and so on.
	regStartMLC := regexp.MustCompile(`(^\/\*.*)`)
	regMiddleMLC := regexp.MustCompile(`.\*.*`)
	regEndMLC := regexp.MustCompile(`.*.(\*/)`)

	// There might be cases when legit string contains "//". Example: ftm.Println("socks5://user...")
	// So we need to consider that.
	regSpecial := regexp.MustCompile(`\"(.*[\t]*\/\/.*)\"`)

	body, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(body), "\n")

	for i, line := range lines {
		if regBasicComments.MatchString(line) {
			if regSpecial.MatchString(line) {
				continue
			}
			lines[i] = strings.TrimSuffix(line, regBasicComments.FindString(line)) // update modified line.
			continue
		}
		if regStartMLC.MatchString(line) {
			lines[i] = strings.TrimPrefix(line, regStartMLC.FindString(line)) // update modified line.
			blockStarted = true
			continue
		}
		if blockStarted {
			if regMiddleMLC.MatchString(line) {
				lines[i] = strings.TrimPrefix(line, regMiddleMLC.FindString(line)) // update modified line.
				continue
			}
			if regEndMLC.MatchString(line) {
				lines[i] = strings.TrimPrefix(line, regEndMLC.FindString(line)) // update modified line.
				blockStarted = false
				continue
			}
		}


	}

	output := strings.Join(lines, "\n")

	return output, nil
}

func walkingDead(src, outputPath string) {

	// Check whether given object is dir or file
	object, err := os.Open(src)
	if err != nil {
		fmt.Println(err)
		return
	}
	info, err := object.Stat()
	if err != nil {
		fmt.Println(err)
	}

	// If dir then make recursive listing of files.
	if info.Mode().IsDir() {

		err = filepath.Walk(src,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					fmt.Println(err)
				}

				ext := regexp.MustCompile(`(\.go)|(\.java)|(\.cpp)|(\.c)`)
				mathc := ext.MatchString(info.Name())

				if info.Mode().IsRegular() && mathc {

					newFile, err := cutThis(path)
					if err != nil {
						fmt.Println(err)
					}

					// Make output path
					newPath := outputPath + strings.TrimPrefix(filepath.Dir(path), filepath.VolumeName(path))
					err = dirInfo(newPath)
					if err != nil {
						err = os.MkdirAll(newPath, 0644)
						if err != nil {
							fmt.Println(err)
						}
					}
					err = ioutil.WriteFile(newPath + "\\" +info.Name(), []byte(newFile), 0644)
					if err != nil {
						fmt.Println(err)
					}
				}

				return nil
			})

		if err != nil {
			log.Println(err)
		}
	} else {
		newFile, err := cutThis(outputPath)
		if err != nil {
			fmt.Println(err)
		}

		err = dirInfo(outputPath)
		if err != nil {
			err = os.MkdirAll(outputPath, 0644)
			if err != nil {
				fmt.Println(err)
			}
		}

		err = ioutil.WriteFile(outputPath + "\\" +info.Name(), []byte(newFile), 0644)
		if err != nil {
			fmt.Println(err)
		}

	}

}

// dirInfo checks output dir existence.
func dirInfo(path string) error  {

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("file does not exist")
	}

	return nil
}

func main() {

	Src := flag.String("src", "", "Path to a directory with files \n or to a single file")
	Output := flag.String("out","", "Path to output directory")
	flag.Parse()

	if *Src == " "{
		fmt.Println("src flag cannot be empty")
		flag.PrintDefaults()
		os.Exit(1)
	}

	walkingDead(*Src,  *Output)
}
