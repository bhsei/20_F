package module

import (
	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/git"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/notification/base"
	"code.gitea.io/gitea/services/module"
	"fmt"
)

type (
	notificationService struct {
		base.NullNotifier
		issueQueue chan issueNotificationOpts
	}
	issueNotificationOpts struct {
		issueID              int64
		commentID            int64
		notificationAuthorID int64
	}
)

var (
	_ base.Notifier = &notificationService{}
)

func NewNotifier() base.Notifier {
	return &notificationService{
		issueQueue: make(chan issueNotificationOpts, 100),
	}
}

func (ns *notificationService) Run() {
	for opts := range ns.issueQueue {
		if err := sendNotification(opts); err != nil {
			log.Error("Was unable to create issue notification: %v", err)
		}
	}
}

// TODO: generate URL for message
func doSend(issue *models.Issue, user *models.User) {
	msg := fmt.Sprintf("issue: %d, author: %s", issue.ID, user.Name)
	module.SendMessage("Issue", msg, "", []int64{user.ID})
}

func sendNotification(opts issueNotificationOpts) error {
	issue, err := models.GetIssueByID(opts.issueID)
	if err != nil {
		return err
	}
	user, err := models.GetUserByID(opts.notificationAuthorID)
	if err != nil {
		return err
	}
	issueWatches, err := models.GetIssueWatchers(opts.issueID)
	if err != nil {
		return err
	}
	watches, err := models.GetWatchers(issue.RepoID)
	if err != nil {
		return err
	}
	alreadyNotified := make(map[int64]struct{}, len(issueWatches)+len(watches))
	notifyUser := func(userID int64) error {
		if userID == opts.notificationAuthorID {
			return nil
		}
		if _, ok := alreadyNotified[userID]; ok {
			return nil
		}
		alreadyNotified[userID] = struct{}{}
		doSend(issue, user)
		return nil
	}
	for _, issueWatch := range issueWatches {
		if !issueWatch.IsWatching {
			alreadyNotified[issueWatch.UserID] = struct{}{}
			continue
		}
		if err := notifyUser(issueWatch.UserID); err != nil {
			return err
		}
	}
	if err = issue.LoadRepo(); err != nil {
		return err
	}
	for _, watch := range watches {
		issue.Repo.Units = nil
		if issue.IsPull && !issue.Repo.CheckUnitUser(watch.UserID, false, models.UnitTypePullRequests) {
			continue
		}
		if !issue.IsPull && !issue.Repo.CheckUnitUser(watch.UserID, false, models.UnitTypeIssues) {
			continue
		}
		if err := notifyUser(watch.UserID); err != nil {
			return err
		}
	}
	return nil
}

func (ns *notificationService) NotifyCreateIssueComment(doer *models.User, repo *models.Repository,
	issue *models.Issue, comment *models.Comment) {
	var opts = issueNotificationOpts{
		issueID:              issue.ID,
		notificationAuthorID: doer.ID,
	}
	if comment != nil {
		opts.commentID = comment.ID
	}
	ns.issueQueue <- opts
}

func (ns *notificationService) NotifyNewIssue(issue *models.Issue) {
	ns.issueQueue <- issueNotificationOpts{
		issueID:              issue.ID,
		notificationAuthorID: issue.Poster.ID,
	}
}

func (ns *notificationService) NotifyIssueChangeStatus(doer *models.User, issue *models.Issue, actionComment *models.Comment, isClosed bool) {
	ns.issueQueue <- issueNotificationOpts{
		issueID:              issue.ID,
		notificationAuthorID: issue.Poster.ID,
	}
}

func (ns *notificationService) NotifyNewPullRequest(pr *models.PullRequest) {
	ns.issueQueue <- issueNotificationOpts{
		issueID:              pr.Issue.ID,
		notificationAuthorID: pr.Issue.PosterID,
	}
}

func (ns *notificationService) NotifyPullRequestReview(pr *models.PullRequest, r *models.Review, comment *models.Comment) {
	var opts = issueNotificationOpts{
		issueID:              pr.Issue.ID,
		notificationAuthorID: r.Reviewer.ID,
	}
	if comment != nil {
		opts.commentID = comment.ID
	}
	ns.issueQueue <- opts
}

func (ns *notificationService) NotifyIssueChangeAssignee(doer *models.User, issue *models.Issue, assignee *models.User, removed bool, comment *models.Comment) {
	ns.issueQueue <- issueNotificationOpts{
		issueID:              issue.ID,
		notificationAuthorID: doer.ID,
	}
}

func (ns *notificationService) NotifyMergePullRequest(pr *models.PullRequest, doer *models.User, baseRepo *git.Repository) {
	ns.issueQueue <- issueNotificationOpts{
		issueID:              pr.Issue.ID,
		notificationAuthorID: doer.ID,
	}
}
