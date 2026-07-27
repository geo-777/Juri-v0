package languages

import "juri/internals/constants"

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
