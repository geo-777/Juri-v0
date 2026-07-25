package compiler

//interface that determines contract.
//can be reused if i later switch to nsjails
import (
	"juri/internals/constants"
	"juri/internals/executor"
)

type Compiler interface {
	Compile(language constants.Language, sourceCode string) (*executor.WorkspaceArtifact, error)
}
