package createprocessing

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"lmcli/processors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)


const templateUrl = "https://raw.githubusercontent.com/mateuszmierzwinski/libmicro-templates/master/%s.%s.template"
const currentSupportedTemplateVersion = "v1"
var configProviders = map[uint8]string{ 1: "cmdconfigprovider\t- Command line given parameters", 2: "envconfigprovider\t- Environment variables given parameters", 3: "yamlconfigprovider\t- Yaml file given parameters (Configuration files)" }


type processing struct {

}

func (b *processing) ProcessCmd(cmd []string) {
	if len(cmd) < 2 {
		b.help()
		return
	}

	switch strings.ToLower(cmd[0]) {
	case "cp":
		b.createProject(cmd[1])
		break
	case "cs":
		if len(cmd) < 3 {
			b.help()
			break
		}
		b.createModule(cmd[1], cmd[2])
		break
	default:
		b.help()
		break
	}
}

func yesNo(wantStr string) bool {
	fmt.Printf("\n:> Do you want to proceed with %s? [Yes | No] (Default: Yes): ", wantStr)
	res := ""
	fmt.Scanf("%s", &res)

	fmt.Println("")

	if len(res) == 0 {
		return true
	}

	return strings.HasPrefix(strings.ToLower(res), "y")
}

func utilCall(utilExecName, utilName, wd string, params ...string) {
	errPipe := bytes.NewBuffer([]byte{})
	stdPipe := bytes.NewBuffer([]byte{})
	cmd := exec.Command(utilExecName, params...)
	cmd.Dir = wd
	cmd.Stderr = errPipe
	cmd.Stdout = stdPipe
	cmd.Env = os.Environ()
	err := cmd.Run()
	if err != nil {
		fmt.Printf("--------- [ ERROR Executing %s ] ----------\n", utilName)
		fmt.Println(stdPipe.String())
		fmt.Printf("-------- [ ERROR Output from %s ] ---------\n", utilName)
		fmt.Println(errPipe.String())
		os.Exit(-1)
	}
}

func goExec(wd string, params ...string) {
	utilCall("go", "GO Compiler", wd, params...)
}

func gitExec(wd string, params ...string) {
	utilCall("git", "Git Tool", wd, params...)
}

func gitIgnoreFileAdd(wd string) {
	buff := bytes.NewBuffer([]byte{})
	buff.WriteString("# IDE files\n.vcs\n.idea\n\n")
	buff.WriteString("# Vendoring\nvendor\n\n")
	buff.WriteString("# OSX Junk files\n.DS_Store\n\n")

	ioutil.WriteFile(filepath.Join(wd, ".gitingore"), buff.Bytes(), os.ModePerm)
}

func sonarProjectFileAdd(wd, projectName string) {
	buff := bytes.NewBuffer([]byte{})
	buff.WriteString(fmt.Sprintf("sonar.projectKey=com.libmicro.%s\n", strings.ToLower(projectName)))
	buff.WriteString("sonar.projectVersion=latest\n")
	buff.WriteString(fmt.Sprintf("ssonar.projectName=%s Micro Service\n", strings.Title(projectName)))
	buff.WriteString("sonar.sources=.\nsonar.language=go\nsonar.sourceEncoding=UTF-8\n")
	buff.WriteString("sonar.coverage.exclusions=**/*_test.go,**/vendor/**\nsonar.exclusions=**/*_test.go,**/vendor/**\n")
	buff.WriteString("sonar.tests=.\nsonar.test.inclusions=**/*_test.go\nsonar.test.exclusions=**/vendor/**\n")
	buff.WriteString("sonar.go.coverage.reportPaths=target/coverage.out\nsonar.go.tests.reportPaths=target/tests.json\n")

	ioutil.WriteFile(filepath.Join(wd, "sonar-project.properties"), buff.Bytes(), os.ModePerm)
}

func gitTemplatePullFile(version, fileToPull string) *[]byte {
	cl := &http.Client{}

	req,err := http.NewRequest("GET", fmt.Sprintf(templateUrl, strings.ToLower(version), strings.ToLower(fileToPull)), nil)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	content,err := cl.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	if content.StatusCode != http.StatusOK {
		fmt.Println("Status code is not 200! Is:", content.StatusCode, content.Status)
		os.Exit(-1)
	}

	data,err := ioutil.ReadAll(content.Body)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	return &data
}

func (b *processing) createProject(cmd string) {
	cnt := gitTemplatePullFile(currentSupportedTemplateVersion, "main.go")
	content := string(*cnt)
	cfgProvider := strings.TrimSpace(selectConfigProvider())

	content = strings.ReplaceAll(content, "{{configProvider}}", cfgProvider)

	wd,err := os.Getwd()
	if err != nil {
		log.Println("Cannot get working directory. Exiting")
		os.Exit(-1)
	}

	projectDir := filepath.Join(wd, cmd)
	log.Println("Project directory will be:", projectDir)
	os.MkdirAll(projectDir, os.ModePerm)

	mainGoFile := filepath.Join(projectDir, "main.go")
	err = ioutil.WriteFile(mainGoFile, []byte(content), os.ModePerm)
	if err != nil {
		log.Println("Cannot create project content:", err.Error())
		os.Exit(-1)
	}

	log.Println("Initializing Go Modules")
	goExec(projectDir, "mod", "init", cmd)

	log.Println("Pulling Go Modules")
	goExec(projectDir,"mod", "tidy")

	if yesNo("Venoring Go Modules") {
		log.Println("Vendoring Go Modules")
		goExec(projectDir,"mod", "vendor")
	}

	log.Println("Trying to build")
	goExec(projectDir,"build", "-o", ".testexec", ".")

	os.Remove(filepath.Join(projectDir, ".testexec"))
	log.Println(";) Build successful")

	if yesNo("SonarQube Scanner integration") {
		log.Println("Adding Sonar-Project.properties file")
		sonarProjectFileAdd(projectDir, cmd)
	}

	if yesNo("Initialize GIT repository") {
		log.Println("Initializing GIT repository")
		gitExec(projectDir, "init")

		log.Println("Adding GIT Ingore file")
		gitIgnoreFileAdd(projectDir)

		if yesNo("First initial Git commit") {
			log.Println("Making GIT initial commit")
			gitExec(projectDir, "add", "--all")
			gitExec(projectDir, "commit", "-m \"Initial commit\"")
		}
	}

	log.Println("OK ;). Ready.")
	exec.Command("open", projectDir).Start()
}

func selectConfigProvider() string {
	fmt.Println("\nSelect configuration provider\n=====================================")
	fmt.Println("")
	for i:=uint8(1); i<=uint8(len(configProviders)); i++ {
		fmt.Printf("\t%d: %s\n", i, configProviders[i])
	}
	fmt.Println("")

	var selected uint8
	fmt.Print("Select config provider (ctrl+c to cancel): ")
	_, err := fmt.Scanf("%d", &selected)
	if err != nil || selected < 1 || selected > uint8(len(configProviders)) {
		log.Println("Unknown config provider selected. Exiting")
		return selectConfigProvider()
	}

	fmt.Println("")
	return strings.Split(configProviders[selected], "\t-")[0]
}

func (b *processing) createModule(proName, modName  string) {
	cnt := gitTemplatePullFile(currentSupportedTemplateVersion, "provider.go")
	log.Println(string(*cnt))
}

func (b *processing) help() {
	fmt.Println("LibMicroCMD LibMicro Command Line Interface")
	fmt.Println("Usage:\tlmcli create project <projectName>")
	fmt.Println("\tlmcli cp <projectName>")
	fmt.Println("\tlmcli cs <projectName> <moduleName>")
	fmt.Println("")

	os.Exit(-1)
}

func New() processors.CmdProcessor {
	return new(processing)
}
