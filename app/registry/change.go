	"ddm-admin-console/service/gerrit"
	"errors"
	currentRevision    = "current"
	mergeList          = "/MERGE_LIST"
	updateMRRetryCount = 5
		return nil, fmt.Errorf("unable to abandon change, %w", err)
	mr, err := a.updateMRStatus(ctx, changeID, "ABANDONED")
	if err != nil {
		return nil, fmt.Errorf("unable to change MR status, %w", err)
	}

	if err := ClearRepoFiles(mr.Spec.ProjectName, a.Cache); err != nil {
		return nil, fmt.Errorf("unable to clear cached files")
		return nil, fmt.Errorf("unable to approve change, %w", err)
	if _, err := a.updateMRStatus(ctx, changeID, "MERGED"); err != nil {
		return nil, fmt.Errorf("unable to change MR status, %w", err)
func (a *App) updateMRStatus(ctx context.Context, changeID, status string) (*gerrit.GerritMergeRequest, error) {
	var mr *gerrit.GerritMergeRequest

	for i := 0; i < updateMRRetryCount; i++ {
		var err error
		mr, err = a.Gerrit.GetMergeRequestByChangeID(ctx, changeID)
			return nil, fmt.Errorf("unable to get MR, %w", err)
			return nil, fmt.Errorf("unable to update MR status, %w", err)

		break
	return mr, nil
		return nil, fmt.Errorf("unable to get gerrit change details, %w", err)
		return nil, fmt.Errorf("unable to get changes, %w", err)
	rspParams := gin.H{
	}

	templateArgs, err := json.Marshal(rspParams)
	if err != nil {
		return nil, errors.New("unable to encode template arguments")
	}

	return router.MakeHTMLResponse(200, "registry/change.html", gin.H{
		"page":         "registry",
		"templateArgs": string(templateArgs),
		return "", fmt.Errorf("unable to get change files, %w", err)
			return "", fmt.Errorf("unable to get file changes, %w", err)
		return "", fmt.Errorf("unable to encode changes, %w", err)
// TODO: split function
		return "", fmt.Errorf("unable to get change commit, %w", err)
		return "", fmt.Errorf("unable to get file content, %w", err)
			return "", fmt.Errorf("unable to create folder, %w", err)
			return "", fmt.Errorf("unable to creaete file, %w", err)
			return "", fmt.Errorf("unable to write string, %w", err)
			return "", fmt.Errorf("unable to close file, %w", err)
		return "", fmt.Errorf("unable to get file content, %w", err)
			return "", fmt.Errorf("unable to create folder, %w", err)
			return "", fmt.Errorf("unable to create file, %w", err)
			return "", fmt.Errorf("unable to write string, %w", err)
			return "", fmt.Errorf("unable to close file, %w", err)