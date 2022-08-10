package registry

import (
	"ddm-admin-console/router"
	"log"
	"net/url"
	"os"
	"os/exec"

	goGerrit "github.com/andygrunwald/go-gerrit"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const (
	currentRevision = "current"
)

func (a *App) viewChange(ctx *gin.Context) (response *router.Response, retErr error) {
	changeID := ctx.Param("change")

	changes, err := a.getChangeContents(changeID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get changes")
	}

	return router.MakeResponse(200, "registry/change.html", gin.H{
		"page":    "registry",
		"changes": changes,
	}), nil
}

func (a *App) getChangeContents(changeID string) (map[string]string, error) {
	changeInfo, _, err := a.Gerrit.GoGerritClient().Changes.GetChangeDetail(changeID, &goGerrit.ChangeOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get gerrit change details")
	}

	files, _, err := a.Gerrit.GoGerritClient().Changes.ListFiles(changeID, currentRevision, &goGerrit.FilesOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get change files")
	}

	//dmp := diffmatchpatch.New()
	changes := make(map[string]string)
	for fileName, fileInfo := range files {
		if fileName == "/COMMIT_MSG" {
			continue
		}

		log.Println(fileInfo)
		originalContent, _, err := a.Gerrit.GoGerritClient().Projects.GetBranchContent(changeInfo.Project, changeInfo.Branch, url.PathEscape(fileName))
		if err != nil {
			return nil, errors.Wrap(err, "unable to get file content")
		}

		f1, err := os.Create("/tmp/f1")
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		f1.WriteString(originalContent)
		f1.Close()

		content, _, err := a.Gerrit.GoGerritClient().Changes.GetContent(changeID, currentRevision, fileName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get file content")
		}

		f2, err := os.Create("/tmp/f2")
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		f2.WriteString(*content)
		f2.Close()

		out, err := exec.Command("diff", "--git", "/tmp/f1", "/tmp/f2").CombinedOutput()
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		changes[fileName] = string(out)

		//log.Println(*content)

		//log.Println(originalContent)
		//dmp.DiffText1()
		//changes[fileName] = template.HTML(dmp.DiffPrettyHtml(dmp.DiffMain(originalContent, *content, false)))
		//log.Println(pretty)
	}

	return changes, nil
}
