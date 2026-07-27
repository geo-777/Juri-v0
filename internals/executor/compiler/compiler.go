package compiler

//interface that determines contract.
//can be reused if i later switch to nsjails
import (
	"context"
	"juri/internals/constants"
	"juri/internals/executor"
)

type Compiler interface {
	Compile(ctx context.Context, language constants.Language, sourceCode string) (*executor.WorkspaceArtifact, error)
}
