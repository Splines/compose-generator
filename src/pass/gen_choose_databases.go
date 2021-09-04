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
	if config.FromFile {
		// Generate from config file
		selectedServiceConfigs := GetServiceConfigurationsByName(config, model.TemplateTypeDatabase)
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
					// Add template to selected templates
					selected.DatabaseServices = append(selected.DatabaseServices, template)
					break
				}
			}
		}
	} else {
		// Generate from user input
		availableDatabases := available.DatabaseServices
		items := TemplateListToLabelList(availableDatabases)
		itemsPreselected := TemplateListToPreselectedLabelList(availableDatabases, selected)
		templateSelections := MultiSelectMenuQuestionIndex("Which database services do you need?", items, itemsPreselected)
		for _, index := range templateSelections {
			Pel()
			// Get selected template config
			selectedConfig := available.DatabaseServices[index]
			// Ask questions to the user
			AskTemplateQuestions(project, &selectedConfig)
			// Ask volume questions to the user
			AskForCustomVolumePaths(project, &selectedConfig)
			// Save template to the selected templates
			selected.DatabaseServices = append(selected.DatabaseServices, selectedConfig)
		}
	}
}
