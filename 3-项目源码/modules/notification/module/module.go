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
		title                string
		content              string
		url                  string
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
func doSend(issue *models.Issue, user *models.User, userID int64, title string, content string, url string) {
	msg_title := fmt.Sprintf("%s issue#%d author %s: %s", issue.Repo.FullName(), issue.Index, user.Name, title)
	module.SendMessage(msg_title, content, url, []int64{userID})
}

func sendNotification(opts issueNotificationOpts) error {
	issue, err := models.GetIssueByID(opts.issueID)
	if err != nil {
		return err
	}
	if err = issue.LoadRepo(); err != nil {
		return err
	}
	if err = issue.LoadPoster(); err != nil {
		return err
	}
	if err = issue.LoadPullRequest(); err != nil {
		return err
	}
	unfiltered := make([]int64, 1, 64)
	unfiltered[0] = issue.PosterID
	ids, err := models.GetAssigneeIDsByIssue(issue.ID)
	if err != nil {
		return err
	}
	unfiltered = append(unfiltered, ids...)
	ids, err = models.GetParticipantsIDsByIssueID(issue.ID)
	if err != nil {
		return err
	}
	unfiltered = append(unfiltered, ids...)
	ids, err = models.GetIssueWatchersIDs(issue.ID, true)
	if err != nil {
		return err
	}
	unfiltered = append(unfiltered, ids...)
	ids, err = models.GetRepoWatchersIDs(issue.RepoID)
	if err != nil {
		return err
	}
	unfiltered = append(unfiltered, ids...)
	user, err := models.GetUserByID(opts.notificationAuthorID)
	if err != nil {
		return err
	}
	alreadyNotified := make(map[int64]struct{}, len(unfiltered))
	notifyUser := func(userID int64) error {
		if userID == opts.notificationAuthorID {
			return nil
		}
		if _, ok := alreadyNotified[userID]; ok {
			return nil
		}
		alreadyNotified[userID] = struct{}{}
		doSend(issue, user, userID, opts.title, opts.content, opts.url)
		return nil
	}
	for _, id := range unfiltered {
		if err := notifyUser(id); err != nil {
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
		opts.content = comment.Content
		opts.url = comment.HTMLURL()
	}
	opts.title = "create new comment"
	ns.issueQueue <- opts
}

func (ns *notificationService) NotifyNewIssue(issue *models.Issue) {
	ns.issueQueue <- issueNotificationOpts{
		issueID:              issue.ID,
		notificationAuthorID: issue.Poster.ID,
		title:                "create new issue",
		content:              issue.Title,
		url:                  issue.HTMLURL(),
	}
}

func (ns *notificationService) NotifyIssueChangeStatus(doer *models.User, issue *models.Issue, actionComment *models.Comment, isClosed bool) {
	ns.issueQueue <- issueNotificationOpts{
		issueID:              issue.ID,
		notificationAuthorID: issue.Poster.ID,
		title:                "issue status changed",
		url:                  issue.HTMLURL(),
	}
}

func (ns *notificationService) NotifyNewPullRequest(pr *models.PullRequest) {
	ns.issueQueue <- issueNotificationOpts{
		issueID:              pr.Issue.ID,
		notificationAuthorID: pr.Issue.PosterID,
		title:                "new pull request",
		url:                  pr.Issue.HTMLURL(),
	}
}

func (ns *notificationService) NotifyPullRequestReview(pr *models.PullRequest, r *models.Review, comment *models.Comment) {
	var opts = issueNotificationOpts{
		issueID:              pr.Issue.ID,
		notificationAuthorID: r.Reviewer.ID,
	}
	if comment != nil {
		opts.commentID = comment.ID
		opts.content = comment.Content
		opts.url = comment.HTMLURL()
	}
	opts.title = "new pull request review"
	ns.issueQueue <- opts
}

func (ns *notificationService) NotifyIssueChangeAssignee(doer *models.User, issue *models.Issue, assignee *models.User, removed bool, comment *models.Comment) {
	ns.issueQueue <- issueNotificationOpts{
		issueID:              issue.ID,
		notificationAuthorID: doer.ID,
		title:                "assignee changed",
		url:                  issue.HTMLURL(),
	}
}

func (ns *notificationService) NotifyMergePullRequest(pr *models.PullRequest, doer *models.User, baseRepo *git.Repository) {
	ns.issueQueue <- issueNotificationOpts{
		issueID:              pr.Issue.ID,
		notificationAuthorID: doer.ID,
		title:                "pull request merged",
		url:                  pr.Issue.HTMLURL(),
	}
}
