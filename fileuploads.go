package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// MAX_FILES_IN_FORM is maximum number of files in a form
const MAX_FILES_IN_FORM = 10

// MAX_UPLOAD_SIZE is maximum total size of all files in a form
const MAX_UPLOAD_SIZE = 1048576 * 100 // 100MB

func uploader(r *http.Request, uploadDest string, inputName string) ([]string, error) {

	var fileNamesList = make([]string, 0, MAX_FILES_IN_FORM)

	err := r.ParseMultipartForm(1048576 * 32)
	if err != nil {
		return fileNamesList, err
	}

	uploadingFiles := r.MultipartForm.File[inputName]
	if len(uploadingFiles) == 0 {
		return fileNamesList, nil
	}
	if len(uploadingFiles) > MAX_FILES_IN_FORM {
		return fileNamesList, fmt.Errorf("upload files quantity is too many: %d", len(uploadingFiles))
	}

	for _, fileHeader := range uploadingFiles {

		if fileHeader.Size > MAX_UPLOAD_SIZE {
			return fileNamesList, fmt.Errorf("upload file exceeds max size: %s", fileHeader.Filename)
		}

		srcio, err := fileHeader.Open()
		if err != nil {
			return fileNamesList, err
		}
		defer srcio.Close()

		// buff := make([]byte, 1024)
		// _, err = file.Read(buff)
		// if err != nil {
		// 	return fileNamesList, err
		// }
		// filetype := http.DetectContentType(buff)
		// if filetype != "image/jpeg" && filetype != "image/png" {
		// 	return fileNamesList, err
		// }

		// _, err = file.Seek(0, io.SeekStart)
		// if err != nil {
		// 	return fileNamesList, err
		// }

		if _, err := os.Stat(uploadDest); err != nil {
			if os.IsNotExist(err) {
				err := os.MkdirAll(uploadDest, 0700)
				if err != nil {
					return fileNamesList, err
				}
			}
		}

		ext := filepath.Ext(fileHeader.Filename)
		saveFilename := fileHeader.Filename
		saveDest := filepath.Join(uploadDest, saveFilename)
		for i := 1; fileExists(saveDest); i++ {
			saveFilename = strings.TrimSuffix(fileHeader.Filename, ext) + "-" + strconv.Itoa(i) + ext
			saveDest = filepath.Join(uploadDest, saveFilename)
		}
		dstio, err := os.Create(saveDest)
		if err != nil {
			return fileNamesList, err
		}
		defer dstio.Close()
		_, err = io.Copy(dstio, srcio)
		if err != nil {
			return fileNamesList, err
		}

		fileNamesList = append(fileNamesList, saveFilename)

	}

	return fileNamesList, nil

}

func moveUploadedFilesToFinalDest(origDir string, destDir string, fileList []string) {
	if len(fileList) == 0 {
		return
	}
	if _, err := os.Stat(destDir); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(destDir, 0700)
			if err != nil {
				log.Println(currentFunction()+":", err, "dir:"+destDir)
			}
		}
	}
	for _, fileName := range fileList {
		err := os.Rename(filepath.Join(origDir, fileName), filepath.Join(destDir, fileName))
		if err != nil {
			log.Println(currentFunction()+":", err, "dir:"+destDir, "file:"+fileName)
		}
	}
}

func removeUploadedFiles(dir string, fileList []string) error {
	if len(fileList) == 0 {
		return nil
	}
	if _, err := os.Stat(dir); err != nil {
		log.Println(currentFunction()+":", err, "dir:"+dir)
		return err
	}
	for _, fileName := range fileList {
		err := os.Remove(filepath.Join(dir, fileName))
		if err != nil {
			log.Println(currentFunction()+":", err, "dir:"+dir, "file:"+fileName)
			return err
		}
	}
	return nil
}

func removeUploadedDirs(objectsDir string, ids []int) {
	for _, id := range ids {
		dir := filepath.Join(objectsDir, strconv.Itoa(id))
		if _, err := os.Stat(dir); err != nil {
			if os.IsNotExist(err) {
				return
			} else {
				log.Println(currentFunction()+":", err, "dir:"+dir)
				return
			}
		}
		err := os.RemoveAll(dir)
		if err != nil {
			log.Println(currentFunction()+":", err, "dir:"+dir)
			return
		}
	}
}
