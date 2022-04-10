package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func shellOpen(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("cmd", "/c", "start", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = exec.Command("xdg-open", url).Start()
	}
	if err != nil {
		log.Println(currentFunction()+":", err)
	}
}

func runClient(UseTLS string, addr string, MSecToWait int, retries int) {
	protocol := "http"
	if UseTLS == "true" {
		protocol = "https"
	}
	client := http.Client{Timeout: 1 * time.Second}
	// Before launching a browser we check if the server is running
	for i := 0; i < retries; i++ {
		time.Sleep(time.Duration(MSecToWait) * time.Millisecond)
		resp, err := client.Get(protocol + "://" + addr + "/assets/")
		if err == nil {
			resp.Body.Close()
			break
		}
	}
	shellOpen(protocol + "://" + addr)
}

func isPrevInstanceRunning(addr string) bool {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return true
	}
	defer l.Close()
	return false
}

func isStringASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func firstLetterIndex(r []rune) int {
	ri := 0
	for i := 0; i < len(r); i++ {
		if unicode.IsLetter(r[i]) {
			ri = i
			break
		}
	}
	return ri
}

func getIDfromURL(path string) (id int, err error) {
	arrayPathElems := strings.Split(path, "/")
	id, err = strconv.Atoi(arrayPathElems[len(arrayPathElems)-1])
	if err != nil {
		return id, err
	}
	return id, nil
}

func getTextIDfromURL(path string) string {
	arrayPathElems := strings.Split(path, "/")
	return arrayPathElems[len(arrayPathElems)-1]
}

func throwAccessDenied(w http.ResponseWriter, logmsg string, userID int, resourceID int) {
	log.Printf("Wrong credentials while: %s, user ID:%d, resource ID:%d\n", logmsg, userID, resourceID)
	http.Error(w, "Wrong credentials write or access attempt detected", http.StatusForbidden)
}

func throwServerError(w http.ResponseWriter, logmsg string, userID int, resourceID int) {
	log.Printf("Internal server error while: %s, user ID:%d, resource ID:%d\n", logmsg, userID, resourceID)
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}

func currentFunction() string {
	counter, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(counter).Name()
}

func intSlicesEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func getIPAddr(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func strToInt(s string) (n int) {
	n, _ = strconv.Atoi(s)
	return n
}

func filterSliceInt(srcList []int, rval int) []int {
	var res []int
	for _, ival := range srcList {
		if ival != rval {
			res = append(res, ival)
		}
	}
	return res
}

func sliceContainsInt(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func filterSliceStr(srcList []string, removalList []string) []string {
	var res []string
	for _, fname := range srcList {
		if !sliceContainsStr(removalList, fname) {
			res = append(res, fname)
		}
	}
	return res
}

func sliceContainsStr(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func calcMaxPages(elemsOnPage int, filteredNum int) int {
	result := 1
	q := filteredNum / elemsOnPage
	r := filteredNum % elemsOnPage
	if q >= 1 {
		result = q
	}
	if r > 0 {
		result++
	}
	return result
}

func replaceBBCodeWithHTML(cont string) string {
	cont = strings.ReplaceAll(cont, "[b]", "<b>")
	cont = strings.ReplaceAll(cont, "[/b]", "</b>")
	cont = strings.ReplaceAll(cont, "[i]", "<i>")
	cont = strings.ReplaceAll(cont, "[/i]", "</i>")
	cont = strings.ReplaceAll(cont, "[u]", "<u>")
	cont = strings.ReplaceAll(cont, "[/u]", "</u>")
	cont = strings.ReplaceAll(cont, "[code]", "<pre>")
	cont = strings.ReplaceAll(cont, "[/code]", "</pre>")

	arr := make([]byte, 7, 7)
	insidetagname := false
	outsidecode := true
	arrcounter := 0
	for i := 0; i < len(cont); i++ {
		if cont[i] == '<' {
			arr[0] = '<'
			insidetagname = true
			arrcounter = 0
		}
		if insidetagname {
			arr[arrcounter] = cont[i]
			arrcounter++
			if arrcounter > 6 {
				arrcounter = 1
			}
			if cont[i] == '>' {
				insidetagname = false
				arrcounter = 0
			}
		}
		if string(arr[:6]) == "<code>" || string(arr[:5]) == "<pre>" {
			outsidecode = false
		}
		if string(arr) == "</code>" || string(arr[:6]) == "</pre>" {
			outsidecode = true
		}
		if !insidetagname {
			for j := 0; j < 7; j++ {
				arr[j] = 0
			}
		}
		if outsidecode && cont[i] == '\r' {
			cont = cont[:i] + cont[i+1:]
		}
		if outsidecode && cont[i] == '\n' {
			cont = cont[:i] + "<br>" + cont[i+1:]
			i += 3
		}
	}

	return cont
}
