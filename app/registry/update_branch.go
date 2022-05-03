package registry

import (
	"context"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"ddm-admin-console/service"
	"ddm-admin-console/service/codebase"
	"ddm-admin-console/service/gerrit"
)

type SortByVersion []string

func (a SortByVersion) Len() int           { return len(a) }
func (a SortByVersion) Less(i, j int) bool { return branchVersion(a[i]) < branchVersion(a[j]) }
func (a SortByVersion) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func branchVersion(name string) int {
	nums := regexp.MustCompile(`\d+`)
	matches := nums.FindAllString(name, -1)
	num := strings.Join(matches, "")
	if num == "" {
		return 0
	}

	version, err := strconv.ParseInt(num, 10, 32)
	if err != nil {
		panic(err)
	}

	return int(version)
}

func HasUpdate(ctx context.Context, gerritService gerrit.ServiceInterface, cb *codebase.Codebase) (bool, []string, error) {
	gerritProject, err := gerritService.GetProject(ctx, cb.Name)
	if service.IsErrNotFound(err) {
		return false, []string{}, nil
	}

	if err != nil {
		return false, nil, errors.Wrap(err, "unable to get gerrit project")
	}

	branches := updateBranches(gerritProject.Status.Branches)

	if len(branches) == 0 {
		return false, branches, nil
	}

	registryVersion := branchVersion(cb.Spec.DefaultBranch)
	if cb.Spec.BranchToCopyInDefaultBranch != "" {
		registryVersion = branchVersion(cb.Spec.BranchToCopyInDefaultBranch)
	}

	mrs, err := gerritService.GetMergeRequestByProject(ctx, gerritProject.Spec.Name)
	if err != nil {
		return false, branches, errors.Wrap(err, "unable to get merge requests")
	}

	branchesDict := make(map[string]string)
	for _, br := range branches {
		branchesDict[br] = br
	}

	for _, mr := range mrs {
		if mr.Status.Value == "NEW" {
			return false, branches, nil
		}

		if mr.Status.Value == "MERGED" {
			mergedBranchVersion := branchVersion(mr.Spec.SourceBranch)
			if mergedBranchVersion > registryVersion {
				registryVersion = mergedBranchVersion
			}

			delete(branchesDict, mr.Spec.SourceBranch)
		}
	}

	branches = []string{}
	for _, br := range branchesDict {
		if branchVersion(br) > registryVersion {
			branches = append(branches, br)
		}
	}

	sort.Sort(SortByVersion(branches))
	return true, branches, nil
}

func updateBranches(projectBranches []string) []string {
	var updateBranches []string
	for _, br := range projectBranches {
		if strings.Contains(br, "refs/heads") && !strings.Contains(br, "master") {
			updateBranches = append(updateBranches, strings.Replace(br, "refs/heads/", "", 1))
		}
	}

	return updateBranches
}

func (a *App) filterUpdateBranchesByCluster(ctx context.Context, registryBranches []string) ([]string, error) {
	clusterCodebase, err := a.codebaseService.Get(a.clusterCodebaseName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster codebase")
	}

	hasUpdate, clusterBranches, err := HasUpdate(ctx, a.gerritService, clusterCodebase)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster update branches")
	}

	if !hasUpdate {
		return []string{}, nil
	}

	if len(clusterBranches) == 0 {
		return []string{}, nil
	}

	clusterBranch := clusterBranches[0]

	var filteredBranches []string
	for _, rb := range registryBranches {
		if branchVersion(rb) <= branchVersion(clusterBranch) {
			filteredBranches = append(filteredBranches, rb)
		}
	}

	return filteredBranches, nil
}
