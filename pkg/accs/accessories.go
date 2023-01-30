package accs

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// FilterRender is for rendering frontend for filters in templates
type FilterRender struct {
	Name                 string
	Attri18n             string
	DisplayName          string
	UseCalendarInConrols bool
	Currencies           map[int]string
}

// ReturnFilterRender provides the function to render filter component to run in a template
func ReturnFilterRender(Name string, Attri18n string, DisplayName string, UseCalendarInConrols bool, Currencies map[int]string) FilterRender {
	return FilterRender{Name, Attri18n, DisplayName, UseCalendarInConrols, Currencies}
}

// HeadRender is for rendering frontend for head block in templates
type HeadRender struct {
	AppTitle    string
	PageTitle   string
	LangCode    string
	SystemTheme string
}

// ReturnHeadRender provides the function to render page header to run in a template
func ReturnHeadRender(AppTitle string, PageTitle string, LangCode string, SystemTheme string) HeadRender {
	return HeadRender{AppTitle, PageTitle, LangCode, SystemTheme}
}

// IsThemeSystem checks is filename has prefix "system-"
func IsThemeSystem(themeName string) bool {
	if strings.HasPrefix(themeName, "system-") {
		return true
	}
	return false
}

// FileExists checks file existence
func FileExists(name string) bool {
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
		log.Println(CurrentFunction()+":", err)
	}
}

// RunClient opens a client program for the specified URL
func RunClient(UseTLS string, addr string, MSecToWait int, retries int) {
	protocol := "http"
	if UseTLS == "true" {
		protocol = "https"
	}
	client := http.Client{Timeout: 1 * time.Second}
	// Before launching a client we check if the server is running
	for i := 0; i < retries; i++ {
		time.Sleep(time.Duration(MSecToWait) * time.Millisecond)
		resp, err := client.Get(protocol + "://" + addr + "/static/")
		if err == nil {
			resp.Body.Close()
			break
		}
	}
	shellOpen(protocol + "://" + addr)
}

// IsPrevInstanceRunning checks if a port is free to run the app
func IsPrevInstanceRunning(addr string) bool {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return true
	}
	defer l.Close()
	return false
}

// IsStringASCII cheks ia a string is pure ASCII
func IsStringASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// FirstLetterIndex return the first alphabetic symbol (which is not a digit or other character)
func FirstLetterIndex(r []rune) int {
	ri := 0
	for i := 0; i < len(r); i++ {
		if unicode.IsLetter(r[i]) {
			ri = i
			break
		}
	}
	return ri
}

// GetIDfromURL returns last path element from URL path as int
func GetIDfromURL(path string) (id int, err error) {
	arrayPathElems := strings.Split(path, "/")
	id, err = strconv.Atoi(arrayPathElems[len(arrayPathElems)-1])
	if err != nil {
		return id, err
	} else if len(arrayPathElems) > 1 {
		id, err = strconv.Atoi(arrayPathElems[len(arrayPathElems)-2])
		if err != nil {
			return id, err
		}
		return id, nil
	}
	return id, nil
}

// GetTextIDfromURL returns last path element from URL path as string
func GetTextIDfromURL(path string) string {
	arrayPathElems := strings.Split(path, "/")
	res := arrayPathElems[len(arrayPathElems)-1]
	if res == "approval" {
		res = arrayPathElems[len(arrayPathElems)-2]
	}
	return res
}

func GetAbsoluteOrRelativePath(defaultPath string, targetPath string) string {
	if strings.HasPrefix(targetPath, "/") || strings.HasPrefix(targetPath, "C:") {
		return targetPath
	}
	return filepath.Join(defaultPath, targetPath)
}

// ThrowAccessDenied writes http.StatusForbidden to responce writer and JSON
func ThrowAccessDenied(w http.ResponseWriter, logmsg string, userID int, resourceID int) {
	log.Printf("Wrong credentials while: %s, user ID:%d, resource ID:%d\n", logmsg, userID, resourceID)
	w.WriteHeader(http.StatusForbidden)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Error int `json:"error"`
	}{403})
}

// ThrowObjectNotFound writes http.StatusNotFound to responce writer and JSON is request was in JSON
func ThrowObjectNotFound(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("api") == "json" || r.FormValue("api") == "json" {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			Error int `json:"error"`
		}{404})
	} else {
		http.NotFound(w, r)
	}
}

// ThrowServerError writes http.StatusInternalServerError to responce writer and JSON
func ThrowServerError(w http.ResponseWriter, logmsg string, userID int, resourceID int) {
	log.Printf("Internal server error: %s, user ID:%d, resource ID:%d\n", logmsg, userID, resourceID)
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Error int `json:"error"`
	}{500})
}

// CurrentFunction returns the name of the current function
func CurrentFunction() string {
	counter, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(counter).Name()
}

// IntSlicesEqual compares two slices of ints for equality
func IntSlicesEqual(a, b []int) bool {
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

// GetIPAddr returns IP address from http.Request
func GetIPAddr(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

// StrToInt silently converts a string to int without returning second value
func StrToInt(s string) (n int) {
	n, _ = strconv.Atoi(s)
	return n
}

// FilterSliceInt removes an int from a slice
func FilterSliceInt(srcList []int, rval int) []int {
	var res []int
	for _, ival := range srcList {
		if ival != rval {
			res = append(res, ival)
		}
	}
	return res
}

// SliceContainsInt checks if a slice contains the int
func SliceContainsInt(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// FilterSliceStrList removes strings in removalList from srcList
func FilterSliceStrList(srcList []string, removalList []string) []string {
	var res []string
	for _, fname := range srcList {
		if !SliceContainsStr(removalList, fname) {
			res = append(res, fname)
		}
	}
	return res
}

// SliceContainsStr cheks is a slice contains the string
func SliceContainsStr(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// CalcMaxPages calculate how many pages a pagination has absed on elements on page and total elements
func CalcMaxPages(elemsOnPage int, filteredNum int) int {
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

// ReplaceBBCodeWithHTML replaces bb-code in a string with HTML tags
func ReplaceBBCodeWithHTML(cont string) string {
	cont = strings.ReplaceAll(cont, "[b]", "<b>")
	cont = strings.ReplaceAll(cont, "[/b]", "</b>")
	cont = strings.ReplaceAll(cont, "[i]", "<i>")
	cont = strings.ReplaceAll(cont, "[/i]", "</i>")
	cont = strings.ReplaceAll(cont, "[u]", "<u>")
	cont = strings.ReplaceAll(cont, "[/u]", "</u>")
	cont = strings.ReplaceAll(cont, "[q]", "<q>")
	cont = strings.ReplaceAll(cont, "[/q]", "</q>")
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

// IntToBool silently converts int to bool
func IntToBool(v int) bool {
	if v != 0 {
		return true
	}
	return false
}

// StrToBool silently converts int to bool
func StrToBool(v string) bool {
	res, _ := strconv.ParseBool(v)
	return res
}
