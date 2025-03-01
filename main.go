package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func convertURL(url string) string {
	reg := regexp.MustCompile(`https://github.com/([^/]+)/([^/]+)/blob/(.+)`)
	res := reg.FindStringSubmatch(url)
	if len(res) == 4 {
		return "https://raw.githubusercontent.com/" + res[1] + "/" + res[2] + "/" + res[3]
	}
	return url
}

func getCode(url string) (string, error) {
	url = convertURL(url)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("Ошибка загрузки: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Ошибка: сервер вернул статус %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Ошибка чтения: %v", err)
	}

	if strings.HasPrefix(strings.ToLower(string(body)), "<!doctype") || strings.HasPrefix(strings.ToLower(string(body)), "<html") {
		return "", fmt.Errorf("Ошибка: сервер вернул HTML")
	}

	return string(body), nil
}

func runCode(lang string, code string, args []string) error {
	interp := map[string]string{
		"python": "python3",
		"perl":   "perl",
		"sh":     "bash",
		"ruby":   "ruby",
		"js":     "node",
	}

	prog, ok := interp[lang]
	if !ok {
		return fmt.Errorf("Язык %s не поддерживается", lang)
	}

	temp, err := os.CreateTemp("", "script-*."+lang)
	if err != nil {
		return fmt.Errorf("Ошибка создания файла: %v", err)
	}
	defer os.Remove(temp.Name())

	_, err = temp.WriteString(code)
	if err != nil {
		return fmt.Errorf("Ошибка записи кода: %v", err)
	}
	temp.Close()

	cmd := exec.Command(prog, append([]string{temp.Name()}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Ошибка выполнения: %v", err)
	}

	return nil
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Использование: go run main.go -r <язык> <URL> [аргументы]")
		fmt.Println("Пример: go run main.go -r python https://github.com/user/repo/blob/main/script.py arg1")
		return
	}

	lang := os.Args[2]
	url := os.Args[3]
	args := os.Args[4:]

	code, err := getCode(url)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	fmt.Println("Код загружен, выполняем...")

	err = runCode(lang, code, args)
	if err != nil {
		fmt.Println("Ошибка:", err)
	}
}
