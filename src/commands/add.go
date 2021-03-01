package commands

import (
	"compose-generator/model"
	"compose-generator/utils"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/compose-generator/diu"
	"github.com/fatih/color"
	yaml "gopkg.in/yaml.v3"
)

// ---------------------------------------------------------------- Public functions ---------------------------------------------------------------

// Add a service to an existing compose file
func Add(flagAdvanced bool, flagRun bool, flagDetached bool, flagForce bool) {
	// Ask for custom YAML file
	path := "./docker-compose.yml"
	if flagAdvanced {
		path = utils.TextQuestionWithDefault("Which compose file do you want to add the service to?", "./docker-compose.yml")
	}

	fmt.Print("Parsing compose file ...")
	// Load compose file
	jsonFile, err := os.Open(path)
	if err != nil {
		utils.Error("Internal error - unable to load compose file", true)
	}
	bytes, _ := ioutil.ReadAll(jsonFile)

	// Parse YAML
	composeFile := model.ComposeFile{}
	if err = yaml.Unmarshal(bytes, &composeFile); err != nil {
		utils.Error("Internal error - unable to parse compose file", true)
	}

	serviceNames := []string{}
	for name := range composeFile.Services {
		serviceNames = append(serviceNames, name)
	}
	color.Green(" done")
	fmt.Println()

	// Ask if the image should be built from source
	build := utils.YesNoQuestion("Build from source?", false)
	var buildPath string
	registry := ""
	if build {
		// Ask for build path
		buildPath = utils.TextQuestionWithDefault("Where is your Dockerfile located?", ".")
		// Check if Dockerfile exists
		if !utils.FileExists(buildPath+"/Dockerfile") && !utils.FileExists(buildPath+"Dockerfile") {
			utils.Error("Aborting. The Dockerfile cannot be found.", true)
		}
	} else {
		// Ask for registry
		registry = utils.TextQuestionWithDefault("From which registry do you want to pick?", "docker.io")
		if registry == "docker.io" {
			registry = ""
		} else {
			registry = registry + "/"
		}
	}

	// Ask for image
	imageName := askForImage()

	// Search for remote image and check manifest
	if !build && !flagForce {
		searchRemoteImage(registry, imageName)
	}

	// Ask for service name
	serviceName := askForServiceName(composeFile.Services, imageName)

	// Ask for container name
	containerName := serviceName
	if flagAdvanced {
		containerName = askForContainerName(serviceName)
	}

	// Ask for volumes
	volumes := askForVolumes()

	// Ask for networks
	networks := askForNetworks()

	// Ask for ports
	ports := askForPorts()

	// Ask for env files
	envFiles := askForEnvFiles()

	// Ask for env variables
	envVariables := []string{}
	if len(envFiles) == 0 {
		envVariables = askForEnvVariables()
	}

	// Ask for depends on
	dependsServices := askForDependsOn(utils.RemoveStringFromSlice(serviceNames, serviceName))

	// Ask for restart mode
	restartValue := askForRestart(flagAdvanced)

	// Add service
	fmt.Print("Adding service ...")
	service := model.Service{
		Build:         buildPath,
		Image:         registry + imageName,
		ContainerName: containerName,
		Volumes:       volumes,
		Networks:      networks,
		Ports:         ports,
		Restart:       restartValue,
		DependsOn:     dependsServices,
		EnvFile:       envFiles,
		Environment:   envVariables,
	}
	composeFile.Services[serviceName] = service
	color.Green(" done")

	// Write to file
	fmt.Print("Saving compose file ...")
	output, err1 := yaml.Marshal(&composeFile)
	err2 := ioutil.WriteFile(path, output, 0777)
	if err1 != nil || err2 != nil {
		utils.Error("Could not write yaml to compose file.", true)
	}
	color.Green(" done")

	// Run if the corresponding flag is set
	if flagRun || flagDetached {
		utils.DockerComposeUp(flagDetached)
	}
}

// --------------------------------------------------------------- Private functions ---------------------------------------------------------------

func askForImage() (name string) {
	name = utils.TextQuestionWithDefault("From which image do you want to build your service?", "hello-world")
	return
}

func searchRemoteImage(registry string, image string) {
	fmt.Print("\nSearching image ...")
	manifest, err := diu.GetImageManifest(registry + image)
	if err == nil {
		color.Green(" found (" + strconv.Itoa(len(manifest.SchemaV2Manifest.Layers)) + " layer(s))\n\n")
	} else {
		color.Red(" not found or no access\n")
		proceed := utils.YesNoQuestion("Proceed anyway?", false)
		if !proceed {
			os.Exit(0)
		}
		fmt.Println()
	}
}

func askForServiceName(existingServices map[string]model.Service, imageName string) (name string) {
	// Set image name as default service name
	defaultName := imageName
	i := strings.Index(defaultName, "/")
	if i > -1 {
		defaultName = defaultName[i+1:]
	}
	i = strings.Index(defaultName, ":")
	if i > -1 {
		defaultName = defaultName[:i]
	}

	// Ask for the service name
	name = utils.TextQuestionWithDefault("How do you want to call your service (best practise: lower cased):", defaultName)
	if _, exists := existingServices[name]; exists {
		// Service name already exists
		if !utils.YesNoQuestion("This service name alreay exists in the compose file. It will be overwritten if you continue. Continue?", false) {
			os.Exit(0)
		}
	}
	return
}

func askForContainerName(serviceName string) (name string) {
	name = utils.TextQuestionWithDefault("How do you want to call your container (best practise: lower cased):", serviceName)
	return
}

func askForVolumes() (volumes []string) {
	if utils.YesNoQuestion("Do you want to add volumes to your service?", false) {
		fmt.Println()
	Volumes:
		// Ask user for volume attachments
		globalVolume := utils.YesNoQuestion("Do you want to add an existing global volume (y) or link a directory / file (n)?", false)
		volumeOuter := ""
		if globalVolume {
			globalVolumes, err := diu.GetExistingVolumes()
			if err == nil {
				menuItmes := []string{}
				for _, volume := range globalVolumes {
					menuItmes = append(menuItmes, volume.Name+" | Driver: "+volume.Driver)
				}
				if len(globalVolumes) >= 1 {
					itemIndex := utils.MenuQuestionIndex("Which global volume?", menuItmes)
					volumeOuter = globalVolumes[itemIndex].Name
				} else if utils.YesNoQuestion("No global volumes found. Do you want to create one?", true) {

				}
			} else {
				utils.Error("Error parsing global volumes.", false)
			}
		} else {
			volumeOuter = utils.TextQuestionWithSuggestions("Directory / file on host machine:", func(toComplete string) (files []string) {
				files, _ = filepath.Glob(toComplete + "*")
				return
			})
			volumeOuter = strings.TrimSpace(volumeOuter)
			if !strings.HasPrefix(volumeOuter, "./") && !strings.HasPrefix(volumeOuter, "/") {
				volumeOuter = "./" + volumeOuter
			}
		}
		volumeInner := utils.TextQuestion("Directory / file inside the container:")
		volumes = append(volumes, volumeOuter+":"+volumeInner)
		// Ask for another volume
		if utils.YesNoQuestion("Share another volume?", true) {
			goto Volumes
		}
		fmt.Println()
	}
	return
}

func askForNetworks() (networks []string) {
	if utils.YesNoQuestion("Do you want to add networks to your service?", false) {

	}
	return
}

func askForPorts() (ports []string) {
	if utils.YesNoQuestion("Do you want to expose ports of your service?", false) {
		fmt.Println()
	Ports:
		// Ask user for port exposures
		portInner := utils.TextQuestionWithValidator("Which port do you want to expose? (inner port)", utils.PortValidator)
		portOuter := utils.TextQuestionWithValidator("To which destination port on the host machine?", utils.PortValidator)
		ports = append(ports, portOuter+":"+portInner)
		// Ask for another env file
		if utils.YesNoQuestion("Expose another port?", true) {
			goto Ports
		}
		fmt.Println()
	}
	return
}

func askForEnvVariables() (envVariables []string) {
	if utils.YesNoQuestion("Do you want to provide environment variables to your service?", false) {
		fmt.Println()
	EnvVar:
		// Ask for environment variables
		variableName := utils.TextQuestionWithValidator("Variable name:", utils.EnvVarNameValidator)
		variableValue := utils.TextQuestion("Variable value:")
		envVariables = append(envVariables, variableName+"="+variableValue)
		// Ask for another env file
		if utils.YesNoQuestion("Expose another environment variable?", true) {
			goto EnvVar
		}
		fmt.Println()
	}
	return
}

func askForEnvFiles() (envFiles []string) {
	if utils.YesNoQuestion("Do you want to provide an environment file to your service?", false) {
		fmt.Println()
	EnvFile:
		// Ask user for env file with auto-suggested text input
		envFile := utils.TextQuestionWithDefaultAndSuggestions("Where is your env file located?", "environment.env", func(toComplete string) (files []string) {
			files, _ = filepath.Glob(toComplete + "*.*")
			return
		})
		// Check if the selected file is valid
		if !utils.FileExists(envFile) || utils.IsDirectory(envFile) {
			utils.Error("File is not valid. Please select another file", false)
			goto EnvFile
		}
		envFiles = append(envFiles, envFile)
		// Ask for another env file
		if utils.YesNoQuestion("Add another environment file?", true) {
			goto EnvFile
		}
		fmt.Println()
	}
	return
}

func askForDependsOn(serviceNames []string) (dependsServices []string) {
	if utils.YesNoQuestion("Should your service depend on other services?", false) {
		fmt.Println()
		dependsServices = utils.MultiSelectMenuQuestion("From which services should your service depend?", serviceNames)
		fmt.Println()
	}
	return
}

func askForRestart(flagAdvanced bool) (restartValue string) {
	if flagAdvanced {
		items := []string{"always", "on-failure", "unless-stopped", "no"}
		restartValue = utils.MenuQuestion("When should the service get restarted?", items)
		fmt.Println()
	}
	return
}
