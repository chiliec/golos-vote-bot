package client

import (
	// Stdlib
	"fmt"
	"log"
	"strings"
	"time"

	// Vendor
	"github.com/pkg/errors"
)

//We check whether there is a voter on the list of those who have already voted
func (api *Client) Verify_Voter_Weight(author, permlink, voter string, weight int) bool {
	ans, err := api.Rpc.Database.GetActiveVotes(author, permlink)
	if err != nil {
		log.Println(errors.Wrapf(err, "Error Verify Voter: "))
		return false
	} else {
		for _, v := range ans {
			if v.Voter == voter && v.Percent == weight {
				return true
			}
		}
		return false
	}
}

func (api *Client) Verify_Voter(author, permlink, voter string) bool {
	ans, err := api.Rpc.Database.GetActiveVotes(author, permlink)
	if err != nil {
		log.Println(errors.Wrapf(err, "Error Verify Voter: "))
		return false
	} else {
		for _, v := range ans {
			if v.Voter == voter {
				return true
			}
		}
		return false
	}
}

//We check whether there are voted
func (api *Client) Verify_Votes(author, permlink string) bool {
	ans, err := api.Rpc.Database.GetActiveVotes(author, permlink)
	if err != nil {
		log.Println(errors.Wrapf(err, "Error Verify Votes: "))
		return false
	} else {
		if len(ans) > 0 {
			return true
		} else {
			return false
		}
	}
}

func (api *Client) Verify_Comments(author, permlink string) bool {
	ans, err := api.Rpc.Database.GetContentReplies(author, permlink)
	if err != nil {
		log.Println(errors.Wrapf(err, "Error Verify Comments: "))
		return false
	} else {
		if len(ans) > 0 {
			return true
		} else {
			return false
		}
	}
}

func (api *Client) Verify_Reblogs(author, permlink, rebloger string) bool {
	ans, err := api.Rpc.Follow.GetRebloggedBy(author, permlink)
	if err != nil {
		log.Println(errors.Wrapf(err, "Error Verify Reblogs: "))
		return false
	} else {
		for _, v := range ans {
			if v == rebloger {
				return true
			}
		}
		return false
	}
}

func (api *Client) Verify_Follow(follower, following string) bool {
	ans, err := api.Rpc.Follow.GetFollowing(follower, following, "blog", 1)
	if err != nil {
		log.Println(errors.Wrapf(err, "Error Verify Follow: "))
		return false
	} else {
		for _, v := range ans {
			if (v.Follower == follower) && (v.Following == following) {
				return true
			} else {
				return false
			}
		}
		return false
	}
}

func (api *Client) Verify_Post(author, permlink string) bool {
	ans, err := api.Rpc.Database.GetContent(author, permlink)
	if err != nil {
		log.Println(errors.Wrapf(err, "Error Verify Post: "))
		return false
	} else {
		if (ans.Author == author) && (ans.Permlink == permlink) {
			return true
		} else {
			return false
		}
		return false
	}
}

func (api *Client) Verify_Delegate_Posting_Key_Sign(from_user, to_user string) bool {
	acc, err := api.Rpc.Database.GetAccounts([]string{from_user})
	if err != nil {
		log.Println(errors.Wrapf(err, "Error Verify Delegate Vote Sign: "))
		return false
	} else if len(acc) == 1 {
		for _, v := range acc[0].Posting.AccountAuths {
			tu := strings.Split(strings.Replace(strings.Replace(fmt.Sprintf("%v", v), "[", "", -1), "]", "", -1), " ")
			if tu[0] == to_user {
				if tu[1] == fmt.Sprintf("%v", acc[0].Posting.WeightThreshold) {
					return true
				}
			}
		}
		return false
	} else {
		return false
	}
}

func (api *Client) Verify_First_Post(username string) bool {
	d := time.Now()
	cont, err := api.Rpc.Database.GetDiscussionsByAuthorBeforeDate(username, "", d.Format("2006-01-02T00:00:00"), 100)
	if err != nil {
		log.Println(errors.Wrapf(err, "Error Verify First Post: "))
		return false
	} else {
		if len(cont) > 1 {
			return false
		} else {
			return true
		}
		return false
	}
}

func (api *Client) Verify_User(username string) bool {
	acc, err := api.Rpc.Database.GetAccounts([]string{username})
	if err != nil {
		log.Println(errors.Wrapf(err, "Error Verify User: "))
		return false
	} else if len(acc) == 1 {
		return true
	} else {
		return false
	}
}
