package languages

import "juri/internals/constants"

// RunCommands maps each supported language to the command used for execution.
var RunCommands = map[constants.Language][]string{
	constants.C: {
		"./program",
	},
	constants.CPP: {
		"./program",
	},
	constants.Java: {
		"java",
		"Main",
	},
	constants.Python: {
		"python3",
		"main.py",
	},
}
