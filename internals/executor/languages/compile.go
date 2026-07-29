package languages

import "juri/internals/constants"

// CompileCommands maps each supported language to the command used for compilation.
var CompileCommands = map[constants.Language][]string{
	constants.C: {
		"gcc",
		"main.c",
		"-O2",
		"-o",
		"program",
	},
	constants.CPP: {
		"g++",
		"main.cpp",
		"-O2",
		"-o",
		"program",
	},
	constants.Java: {
		"javac",
		"Main.java",
	},
	constants.Python: nil, // No compilation required
}
