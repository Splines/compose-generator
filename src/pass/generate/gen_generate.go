/*
Copyright © 2021 Compose Generator Contributors
All rights reserved.
*/

package pass

import (
	"compose-generator/model"
	"compose-generator/project"
	"path/filepath"
	"strconv"
)

var generateServiceMockable = generateService

// ---------------------------------------------------------------- Public functions ---------------------------------------------------------------

// Generate transforms the selected templates list and the enriched project to a composition
func Generate(project *model.CGProject, selectedTemplates *model.SelectedTemplates) {
	pel()
	templateCount := selectedTemplates.GetTotal()
	if templateCount > 0 {
		// Generate services from selected templates
		spinner := startProcess("Generating configuration from " + strconv.Itoa(templateCount) + " template(s) ...")

		// Prepare
		if project.WithReadme {
			instructionsHeaderPath := getPredefinedServicesPath() + "/INSTRUCTIONS_HEADER.md"
			project.ReadmeChildPaths = append(project.ReadmeChildPaths, instructionsHeaderPath)
		}

		// Generate frontends
		for _, template := range selectedTemplates.FrontendServices {
			generateServiceMockable(project, selectedTemplates, template, model.TemplateTypeFrontend, template.Name)
		}
		// Generate backends
		for _, template := range selectedTemplates.BackendServices {
			generateServiceMockable(project, selectedTemplates, template, model.TemplateTypeBackend, template.Name)
		}
		// Generate databases
		for _, template := range selectedTemplates.DatabaseServices {
			generateServiceMockable(project, selectedTemplates, template, model.TemplateTypeDatabase, template.Name)
		}
		// Generate db admins
		for _, template := range selectedTemplates.DbAdminServices {
			generateServiceMockable(project, selectedTemplates, template, model.TemplateTypeDbAdmin, template.Name)
		}
		// Generate proxies
		for _, template := range selectedTemplates.ProxyService {
			generateServiceMockable(project, selectedTemplates, template, model.TemplateTypeProxy, template.Name)
		}
		// Generate tls helpers
		for _, template := range selectedTemplates.TlsHelperService {
			generateServiceMockable(project, selectedTemplates, template, model.TemplateTypeTlsHelper, template.Name)
		}
		stopProcess(spinner)
	} else {
		errorLogger.Println("No templates selected. Aborting")
		logError("No templates selected. Aborting ...", true)
	}
}

// --------------------------------------------------------------- Private functions ---------------------------------------------------------------

func generateService(
	proj *model.CGProject,
	selectedTemplates *model.SelectedTemplates,
	template model.PredefinedTemplateConfig,
	templateType string,
	serviceName string,
) {
	infoLogger.Println("Generating service '" + serviceName + "' ...")
	// Load service configuration
	service := loadTemplateService(
		proj,
		selectedTemplates,
		templateType,
		serviceName,
		project.LoadFromDir(template.Dir),
		project.LoadFromComposeFile("service.yml"),
	)
	// Change to build context path to contain more information
	if service.Build != nil && service.Build.Context != "" {
		service.Build.Context = template.Dir + "/" + service.Build.Context
	}
	// Add env variables for proxy questions
	for varName := range proj.ProxyVars[template.Name] {
		varValue := proj.ProxyVars[template.Name][varName]
		service.Environment[varName] = &varValue
	}
	// Add service to the project
	proj.Composition.Services = append(proj.Composition.Services, *service)
	// Add child readme files
	for _, readmePath := range template.GetFilePathsByType(model.FileTypeDocs) {
		proj.ReadmeChildPaths = append(proj.ReadmeChildPaths, filepath.Join(template.Dir, readmePath))
	}
	// Add gitignore patterns
	for _, envFilePath := range template.GetFilePathsByType(model.FileTypeEnv) {
		if !sliceContainsString(proj.GitignorePatterns, envFilePath) {
			proj.GitignorePatterns = append(proj.GitignorePatterns, envFilePath)
		}
	}
	infoLogger.Println("Generating service '" + serviceName + "' (done)")
}
