package process_pair

import(
	"strconv"
	"os/exec"
	"runtime"
	"fmt"
	"os"
)

func StartBackup(id string, port string, pp int) {
	arg_pp := strconv.Itoa(pp)
	currentDir, _ := os.Getwd()

	op_sys := runtime.GOOS
    switch op_sys {
    case "windows":
        exec.Command("cmd", "/C", "start", "powershell -NoExit", "go", "run", "main.go", "--id",id, "--port", port, "--pp", arg_pp).Run()
    case "darwin":
		filename := "main.go"
		cmd := exec.Command("osascript", "-e", `tell app "Terminal" to do script "cd `+currentDir+`; go run `+filename+` --id=`+id+` --port=`+port+` --pp=`+arg_pp+`"`)
		err := cmd.Run()
		if err != nil {
			fmt.Println(err)
		}
    case "linux":
        exec.Command("gnome-terminal", "-x", "go", "run", "main.go", "--id",id, "--port", port, "--pp", arg_pp).Run()
    default:
        fmt.Printf("%s.\n", op_sys)
    }
}