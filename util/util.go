/*
   Provides a multitude of support functions to
   help make development easier. Use of these functions should be prefered,
   as it allows for easier maintainence.
*/
package util

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

/*
   Sends an http request and returns the body. Gives an error if the http request failed
   or returned a non success code.
*/
func HttpRequest(method string, url string, bodyData string) (string, error) {
	//log.Println("URL IS "+url)
	body := strings.NewReader(bodyData)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Println(err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	req.Close = true
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return "", err
	}

	defer resp.Body.Close()
	buf := new(bytes.Buffer)

	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf(buf.String())
	}
	return buf.String(), nil
}

func GetUUIDString() (string, error) {
	uid, err := uuid.NewV4()
	return uid.String(), err
}

/****Basic Linux Functions****/

/*
   Rm removes all of the given directories or files
*/
func Rm(directories ...string) error {
	for _, directory := range directories {
		if conf.Verbose {
			fmt.Printf("Removing  %s...", directory)
		}
		err := os.RemoveAll(directory)
		if conf.Verbose {
			fmt.Printf("done\n")
		}
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

/*
   Mkdir creates a directory
*/
func Mkdir(directory string) error {
	if conf.Verbose {
		fmt.Printf("Creating directory %s\n", directory)
	}
	return os.MkdirAll(directory, 0755)
}

/*
   Cp copies a file
*/
func Cp(src string, dest string) error {
	if conf.Verbose {
		fmt.Printf("Copying %s to %s\n", src, dest)
	}

	cmd := exec.Command("bash", "-c", fmt.Sprintf("cp %s %s", src, dest))
	return cmd.Run()
}

/*
   Cpr copies a directory
*/
func Cpr(src string, dest string) error {
	if conf.Verbose {
		fmt.Printf("Copying %s to %s\n", src, dest)
	}

	cmd := exec.Command("cp", "-r", src, dest)
	return cmd.Run()
}

/*
   Write writes data to a file, creating it if it doesn't exist,
   deleting and recreating it if it does.
*/
func Write(path string, data string) error {
	if conf.Verbose {
		fmt.Printf("Writing to file %s...", path)
	}

	err := ioutil.WriteFile(path, []byte(data), 0664)

	if conf.Verbose {
		fmt.Printf("done\n")
	}
	return err
}

/*
   Lsr lists the contents of a directory recursively
*/
func Lsr(_dir string) ([]string, error) {
	dir := _dir
	if dir[len(dir)-1:] != "/" {
		dir += "/"
	}
	out := []string{}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if f.IsDir() {
			contents, err := Lsr(fmt.Sprintf("%s%s/", dir, f.Name()))
			if err != nil {
				return nil, err
			}
			out = append(out, contents...)
		} else {
			out = append(out, fmt.Sprintf("%s%s", dir, f.Name()))
		}
	}
	return out, nil
}

/*
   LsDir lists directories in order of construction
*/
func LsDir(_dir string) ([]string, error) {
	dir := _dir
	if dir[len(dir)-1:] != "/" {
		dir += "/"
	}
	out := []string{}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	for _, f := range files {
		if f.IsDir() {
			out = append(out, fmt.Sprintf("%s%s/", dir, f.Name()))
			content, err := LsDir(fmt.Sprintf("%s%s/", dir, f.Name()))
			if err != nil {
				log.Println(err)
				return nil, err
			}
			out = append(out, content...)
		}
	}
	return out, nil
}

/*
   CombineConfig combines an Array with \n as the delimiter.
   Useful for generating configuration files.
*/
func CombineConfig(entries []string) string {
	out := ""
	for _, entry := range entries {
		out += fmt.Sprintf("%s\n", entry)
	}
	return out
}

/*
   BashExec executes _cmd in bash then return the result
*/
func BashExec(_cmd string) (string, error) {
	if conf.Verbose {
		fmt.Printf("Executing : %s\n", _cmd)
	}

	cmd := exec.Command("bash", "-c", _cmd)

	var resultsRaw bytes.Buffer

	cmd.Stdout = &resultsRaw
	err := cmd.Start()
	if err != nil {
		log.Println(err)
		return "", err
	}
	err = cmd.Wait()
	if err != nil {
		log.Println(err)
		return "", err
	}

	return resultsRaw.String(), nil
}

/*
   GetPath extracts the base path of the given path
*/
func GetPath(path string) string {
	index := strings.LastIndex(path, "/")
	if index != -1 {
		return path
	}
	return path[:index]
}

/*
   IntArrRemove removes an element from an array of ints
*/
func IntArrRemove(op []int, index int) []int {
	return append(op[:index], op[index+1:]...)
}

/*
   IntArrFill fills the elements of an array according the given
   function, and then returns it.
   f takes in the index and returns the value to place at that index.
*/
func IntArrFill(size int, f func(int) int) []int {
	out := make([]int, size)
	for i := 0; i < size; i++ {
		out[i] = f(i)
	}
	return out
}

/*
   GetJSONNumber checks and extracts a json.Number from data[field].
   Will return an error if data[field] does not exist or is of the wrong type.
*/
func GetJSONNumber(data map[string]interface{}, field string) (json.Number, error) {
	rawValue, exists := data[field]
	if exists && rawValue != nil {
		switch rawValue.(type) {
		case json.Number:
			value, valid := rawValue.(json.Number)
			if !valid {
				return "", fmt.Errorf("Invalid json number")
			}
			return value, nil

		}
	}
	return "", fmt.Errorf("Incorrect type for %s given", field)
}

/*
   GetJSONInt64 checks and extracts a int64 from data[field].
   Will return an error if data[field] does not exist or is of the wrong type.
*/
func GetJSONInt64(data map[string]interface{}, field string, out *int64) error {
	rawValue, exists := data[field]
	if exists && rawValue != nil {
		switch rawValue.(type) {
		case json.Number:
			value, err := rawValue.(json.Number).Int64()
			if err != nil {
				return err
			}
			*out = value
			return nil
		default:
			return fmt.Errorf("Incorrect type for %s given", field)
		}
	}
	return nil
}

/*
   GetJSONInt64 checks and extracts a []string from data[field].
   Will return an error if data[field] does not exist or is of the wrong type.
*/
func GetJSONStringArr(data map[string]interface{}, field string, out *[]string) error {
	rawValue, exists := data[field]
	if exists && rawValue != nil {
		switch rawValue.(type) {
		case []string:
			value, valid := rawValue.([]string)
			if !valid {
				return fmt.Errorf("Invalid string array")
			}
			*out = value
			return nil
		default:
			return fmt.Errorf("Incorrect type for %s given", field)
		}
	}
	return nil
}

/*
   GetJSONInt64 checks and extracts a string from data[field].
   Will return an error if data[field] does not exist or is of the wrong type.
*/
func GetJSONString(data map[string]interface{}, field string, out *string) error {
	rawValue, exists := data[field]
	if exists && rawValue != nil {
		switch rawValue.(type) {
		case string:
			value, valid := rawValue.(string)
			if !valid {
				return fmt.Errorf("Invalid string")
			}
			*out = value
			return nil
		default:
			return fmt.Errorf("Incorrect type for %s given", field)
		}
	}
	return nil
}

/*
   GetJSONInt64 checks and extracts a bool from data[field].
   Will return an error if data[field] does not exist or is of the wrong type.
*/
func GetJSONBool(data map[string]interface{}, field string, out *bool) error {
	rawValue, exists := data[field]
	if exists && rawValue != nil {
		switch rawValue.(type) {
		case bool:
			value, valid := rawValue.(bool)
			if !valid {
				return fmt.Errorf("Invalid bool")
			}
			*out = value
			return nil
		default:
			return fmt.Errorf("Incorrect type for %s given", field)
		}
	}
	return nil
}

func MergeStringMaps(m1 map[string]interface{}, m2 map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k1, v1 := range m1 {
		out[k1] = v1
	}

	for k2, v2 := range m2 {
		out[k2] = v2
	}
	return out
}

func ConvertToStringMap(in interface{}) map[string]string {

	data := in.(map[string]interface{})
	out := make(map[string]string)

	for key, value := range data {
		strval, _ := json.Marshal(value)
		/*switch v := i.(type) {
		    case int:
		        fallthrough
		    case int8:
		        fallthrough
		    case int16:
		        fallthrough
		    case int32:
		        fallthrough
		    case int64:
		        strval = string(strconv.AppendInt(nil, int64(v), 10))

		    case float:
		        fallthrough
		    case float32:
		        fallthrough
		    case float64:
		        b64 = strconv.AppendFloat(nil,float64(v), 'f', -1, 64)
		        fmt.Println("the reciprocal of i is", 1/v)
		    case string:
		        strval = v
		    case []byte:

		    default:
		        // i isn't one of the types above
		}*/
		out[key] = string(strval)
	}
	return out
}

func GetBlockchainConfig(blockchain string, file string, files map[string]string) ([]byte, error) {
	if files != nil {
		res, exists := files["genesis.json"]
		if exists {
			return base64.StdEncoding.DecodeString(res)
		}
	}
	return ioutil.ReadFile(fmt.Sprintf("./resources/%s/%s", blockchain, file))

}
