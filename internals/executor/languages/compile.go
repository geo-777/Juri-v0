package languages

import "juri/internals/constants"

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
