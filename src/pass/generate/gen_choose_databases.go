/*
Copyright © 2021 Compose Generator Contributors
All rights reserved.
*/

package pass

import (
	"compose-generator/model"
)

// ---------------------------------------------------------------- Public functions ---------------------------------------------------------------

// GenerateChooseDatabases lets the user choose predefined database service templates
func GenerateChooseDatabases(
	project *model.CGProject,
	available *model.AvailableTemplates,
	selected *model.SelectedTemplates,
	config *model.GenerateConfig,
) {
	if config != nil && config.FromFile {
		// Generate from config file
		infoLogger.Println("Generating databases from config file ...")
		selectedServiceConfigs := getServiceConfigurationsByType(config, model.TemplateTypeDatabase)
		if project.Vars == nil {
			project.Vars = make(map[string]string)
		}
		for _, template := range available.DatabaseServices {
			for _, selectedConfig := range selectedServiceConfigs {
				if template.Name == selectedConfig.Name {
					// Add vars to project
					for _, question := range template.Questions {
						if value, ok := selectedConfig.Params[question.Variable]; ok {
							project.Vars[question.Variable] = value
						} else {
							project.Vars[question.Variable] = question.DefaultValue
						}
					}
					for _, question := range template.ProxyQuestions {
						if value, ok := selectedConfig.Params[question.Variable]; ok {
							project.Vars[question.Variable] = value
						} else {
							project.Vars[question.Variable] = question.DefaultValue
						}
					}
					for _, question := range template.Volumes {
						if value, ok := selectedConfig.Params[question.Variable]; ok {
							project.Vars[question.Variable] = value
						} else {
							project.Vars[question.Variable] = question.DefaultValue
						}
					}
					// Add template to selected templates
					selected.DatabaseServices = append(selected.DatabaseServices, template)
					break
				}
			}
		}
		infoLogger.Println("Generating databases from config file (done)")
	} else {
		// Generate from user input
		infoLogger.Println("Generating databases from user input ...")
		items := templateListToLabelList(available.DatabaseServices)
		items = append(items, "Custom database service")
		itemsPreselected := templateListToPreselectedLabelList(available.DatabaseServices, selected)
		templateSelections := multiSelectMenuQuestionIndex("Which database services do you need?", items, itemsPreselected)
		for _, index := range templateSelections {
			pel()
			if index == len(available.DatabaseServices) { // Custom service was selected
				GenerateAddCustomService(project, model.TemplateTypeDatabase)
			} else { // Predefined service was selected
				// Get selected template config
				selectedConfig := available.DatabaseServices[index]
				// Ask questions to the user
				askTemplateQuestions(project, &selectedConfig)
				// Ask proxy questions to the user
				askTemplateProxyQuestions(project, &selectedConfig, selected)
				// Ask volume questions to the user
				askForCustomVolumePaths(project, &selectedConfig)
				// Save template to the selected templates
				selected.DatabaseServices = append(selected.DatabaseServices, selectedConfig)
			}
		}
		infoLogger.Println("Generating databases from user input (done)")
	}
}
