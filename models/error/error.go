package error

type CDPipelineExistsError struct {
}

func (e *CDPipelineExistsError) Error() string {
	return "cd pipeline already exists"
}

func NewCDPipelineExistsError() error {
	return &CDPipelineExistsError{}
}

type NonValidRelatedBranchError struct {
}

func (e *NonValidRelatedBranchError) Error() string {
	return "application has non valid related branch"
}

func NewNonValidRelatedBranchError() error {
	return &NonValidRelatedBranchError{}
}

type CDPipelineDoesNotExistError struct {
}

func (e *CDPipelineDoesNotExistError) Error() string {
	return "cd pipeline doesn't exist"
}

func NewCDPipelineDoesNotExistError() error {
	return &CDPipelineDoesNotExistError{}
}

type CodebaseAlreadyExistsError struct {
}

func (e *CodebaseAlreadyExistsError) Error() string {
	return "codebase already exists"
}

func NewCodebaseAlreadyExistsError() error {
	return &CodebaseAlreadyExistsError{}
}

type CodebaseWithGitURLPathAlreadyExistsError struct {
}

func (e *CodebaseWithGitURLPathAlreadyExistsError) Error() string {
	return "codebase with git url path already exists"
}

func NewCodebaseWithGitURLPathAlreadyExistsError() error {
	return &CodebaseWithGitURLPathAlreadyExistsError{}
}
