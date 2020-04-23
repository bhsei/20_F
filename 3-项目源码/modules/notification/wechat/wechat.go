package wechat

import (
	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/git"
	"code.gitea.io/gitea/modules/notification/base"
)

type wechatNotifier struct {
	base.NullNotifier
}

var (
	_ base.Notifier = &wechatNotifier{}
)

func NewNotifier() base.Notifier {
	return &wechatNotifier{}
}

func (w *wechatNotifier) NotifyCreateIssueComment(doer *models.User, repo *models.Repository,
	issue *models.Issue, comment *models.Comment) {
}

func (w *wechatNotifier) NotifyNewIssue(issue *models.Issue) {
}

func (w *wechatNotifier) NotifyIssueChangeStatus(doer *models.User, issue *models.Issue, actionComment *models.Comment, isClosed bool) {
}

func (w *wechatNotifier) NotifyNewPullRequest(pr *models.PullRequest) {
}

func (w *wechatNotifier) NotifyPullRequestReview(pr *models.PullRequest, r *models.Review, comment *models.Comment) {
}

func (w *wechatNotifier) NotifyIssueChangeAssignee(doer *models.User, issue *models.Issue, assignee *models.User, removed bool, comment *models.Comment) {
}

func (w *wechatNotifier) NotifyMergePullRequest(pr *models.PullRequest, doer *models.User, baseRepo *git.Repository) {
}
